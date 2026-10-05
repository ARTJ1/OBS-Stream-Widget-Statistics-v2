package owtracker

import (
	"context"
	"fmt"
	"image"
	"io"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"
)

const (
	StateInactive int32 = 0
	StateMenu     int32 = 1
	StateMatch    int32 = 2
)

const (
	pollDisabled   = 15 * time.Second // auto off — minimal wakeups
	pollNoGame     = 12 * time.Second // auto on, Overwatch not running
	pollInactive   = 3 * time.Second  // game running, not focused
	pollMenu       = 1500 * time.Millisecond
	pollMatch      = 1000 * time.Millisecond
	pollEndScreen  = 400 * time.Millisecond
	pollTest       = 800 * time.Millisecond
	pollTestSlow   = 1200 * time.Millisecond
	pollTestFast   = 400 * time.Millisecond
	antiBlink      = 1 * time.Second
	cooldown       = 40 * time.Second
)

type Outcome string

const (
	OutcomeWin  Outcome = "win"
	OutcomeLoss Outcome = "loss"
)

type Status struct {
	Enabled     bool   `json:"enabled"`
	State       string `json:"state"`
	WindowTitle string `json:"windowTitle"`
	WinReady    bool   `json:"winTemplateReady"`
	LossReady   bool   `json:"lossTemplateReady"`
	WinCustom   bool   `json:"winCustom,omitempty"`
	LossCustom  bool   `json:"lossCustom,omitempty"`
	ZoneCustom  bool   `json:"zoneCustom,omitempty"`
	WinQuality  int    `json:"winQuality,omitempty"`
	LossQuality int    `json:"lossQuality,omitempty"`
	GameRunning bool   `json:"gameRunning"`
	GameFocused bool   `json:"gameFocused"`
	ZoneReady   bool   `json:"zoneReady"`
	Zone        Zone   `json:"zone,omitempty"`
	OcrReady    bool   `json:"ocrReady"`
	DevMode     bool   `json:"devMode,omitempty"`

	CaptureVia  string      `json:"captureVia,omitempty"`  // "obs" | "screen" (last frame source)
	WinOCRLangs []string    `json:"winOcrLangs,omitempty"` // empty until the game first runs
	WinOCRError string      `json:"winOcrError,omitempty"`
	LastResult  *AutoResult `json:"lastResult,omitempty"`
	LastRead    *BannerRead `json:"lastRead,omitempty"` // dev: last frame analysis
}

type Tracker struct {
	dataDir  string
	hashes   hashStore
	enabled  atomic.Bool
	state    atomic.Int32
	onResult func(Outcome)

	lastTitle atomic.Value // string
	winQ      atomic.Int32
	lossQ     atomic.Int32

	testMode       atomic.Bool
	debugLogOn     atomic.Bool
	dryRun         atomic.Bool
	matchThresholdVal atomic.Int32
	captureSource     atomic.Value // string
	debug             *debugLog
	zoneMu            sync.RWMutex
	zone              Zone

	lastAppliedMu sync.Mutex
	lastApplied   *ProbeResult

	tplMu      sync.RWMutex
	winTplImg  image.Image
	lossTplImg image.Image

	lastPeekMu sync.Mutex
	lastPeek   ProbeResult

	lastProbeImgMu sync.Mutex
	lastProbeImg   image.Image

	endScreen endScreenWatch
	bannerStable bannerStability
	autoLearn    *mlAutoStore

	// Production detector: Windows OCR + once-per-match flow (autorun.go).
	ocr        *winOCR
	autoMu     sync.Mutex
	flow       autoFlow
	lastRead   BannerRead
	lastReadAt time.Time
	lastResult *AutoResult
	// nextFullRead throttles stage-2 OCR when the strip fires on scenery.
	nextFullRead time.Time
	cnn          *BannerCNN // embedded banner classifier (nil = auto unavailable)
	vote         cnnVote
	frameSrc     FrameSource  // OBS frames (primary); guarded by autoMu
	obsRetryAt   time.Time    // after an OBS failure, use screen capture until then
	captureVia   atomic.Value // string: "obs" | "screen" | ""
}

func New(dataDir string, onResult func(Outcome)) *Tracker {
	t := &Tracker{dataDir: dataDir, onResult: onResult}
	t.lastTitle.Store("")
	bindOCRDataDir(dataDir)
	t.debug = newDebugLog(dataDir)
	ml.InitText(dataDir)
	t.autoLearn = newMLAutoStore(dataDir)
	t.ocr = newWinOCR(dataDir)
	if m, err := loadEmbeddedBannerCNN(); err != nil {
		log.Printf("owtracker: banner model: %v", err)
	} else {
		t.cnn = m
	}
	t.enabled.Store(loadAutoEnabled(dataDir))
	t.captureSource.Store(CaptureWindow)
	t.applyDebugSettings(loadDebugSettings(dataDir))
	if !DevMode() {
		t.testMode.Store(false)
		t.debugLogOn.Store(false)
		t.dryRun.Store(false)
	}
	t.applySavedConfig(loadSavedHashes(dataDir))
	t.reloadTemplateImages()
	return t
}

func (t *Tracker) applySavedConfig(saved savedHashes) {
	cfg := effectiveConfig(normalizeSaved(saved))
	t.hashes.setWin(cfg.Win)
	t.hashes.setLoss(cfg.Loss)
	t.winQ.Store(int32(cfg.WinQuality))
	t.lossQ.Store(int32(cfg.LossQuality))
	mt := cfg.MatchThreshold
	if mt < 5 || mt > 32 {
		mt = defaultMatchThreshold
	}
	t.matchThresholdVal.Store(int32(mt))
	if cfg.Zone.Valid() {
		t.zone = cfg.Zone
	}
}

func normalizeSaved(saved savedHashes) savedHashes {
	if saved.Win != "" {
		saved.CustomWin = true
	}
	if saved.Loss != "" {
		saved.CustomLoss = true
	}
	if saved.Zone.Valid() {
		saved.CustomZone = true
	}
	return saved
}

func (t *Tracker) getZone() Zone {
	t.zoneMu.RLock()
	defer t.zoneMu.RUnlock()
	return t.zone
}

func (t *Tracker) setZone(z Zone) {
	t.zoneMu.Lock()
	t.zone = z
	t.zoneMu.Unlock()
}

func (t *Tracker) zoneReady() bool {
	return t.getZone().Valid()
}

func (t *Tracker) saveZone(z Zone) error {
	if !z.Valid() {
		return ErrZoneInvalid
	}
	t.setZone(z)
	saved := normalizeSaved(loadSavedHashes(t.dataDir))
	saved.Zone = z
	saved.CustomZone = true
	return saveHashes(t.dataDir, saved)
}

func (t *Tracker) matchThreshold() int {
	v := int(t.matchThresholdVal.Load())
	if v < 5 {
		return defaultMatchThreshold
	}
	return v
}

func (t *Tracker) getCaptureSource() string {
	v, _ := t.captureSource.Load().(string)
	if v == CaptureScreen || v == CaptureWindow {
		return v
	}
	return CaptureBoth
}

func (t *Tracker) SetCaptureSource(v string) {
	switch v {
	case CaptureScreen, CaptureWindow, CaptureBoth:
		t.captureSource.Store(v)
	default:
		t.captureSource.Store(CaptureWindow)
	}
	t.persistDebugSettings()
}

func (t *Tracker) SetMatchThreshold(v int) {
	if v < 5 {
		v = 5
	}
	if v > 32 {
		v = 32
	}
	t.matchThresholdVal.Store(int32(v))
	saved := loadSavedHashes(t.dataDir)
	saved.MatchThreshold = v
	_ = saveHashes(t.dataDir, saved)
}

func (t *Tracker) SetEnabled(v bool) {
	t.enabled.Store(v)
	saveAutoEnabled(t.dataDir, v)
	if !v {
		t.state.Store(StateInactive)
		t.endScreen.reset()
		t.bannerStable.reset()
		t.ocr.Close()
	}
}

func (t *Tracker) Enabled() bool { return t.enabled.Load() }

func (t *Tracker) Toggle() bool {
	for {
		cur := t.enabled.Load()
		next := !cur
		if t.enabled.CompareAndSwap(cur, next) {
			saveAutoEnabled(t.dataDir, next)
			if !next {
				t.state.Store(StateInactive)
				t.endScreen.reset()
				t.bannerStable.reset()
				t.ocr.Close()
			}
			return next
		}
	}
}

func (t *Tracker) StateName() string {
	switch t.state.Load() {
	case StateMenu:
		return "menu"
	case StateMatch:
		return "match"
	default:
		return "inactive"
	}
}

func (t *Tracker) Status() Status {
	win, loss := t.hashes.get()
	saved := effectiveConfig(normalizeSaved(loadSavedHashes(t.dataDir)))
	title := activeWindowTitle()
	t.lastTitle.Store(title)
	t.autoMu.Lock()
	lastResult := t.lastResult
	var lastRead *BannerRead
	if !t.lastReadAt.IsZero() && time.Since(t.lastReadAt) < 10*time.Second {
		r := t.lastRead
		lastRead = &r
	}
	t.autoMu.Unlock()
	via, _ := t.captureVia.Load().(string)
	return Status{
		CaptureVia:  via,
		WinOCRLangs: t.ocr.Langs(),
		WinOCRError: t.ocr.LastError(),
		LastResult:  lastResult,
		LastRead:    lastRead,
		Enabled:     t.Enabled(),
		State:       t.StateName(),
		WindowTitle: title,
		WinReady:    win != "",
		LossReady:   loss != "",
		WinCustom:   saved.CustomWin,
		LossCustom:  saved.CustomLoss,
		ZoneCustom:  saved.CustomZone,
		WinQuality:  int(t.winQ.Load()),
		LossQuality: int(t.lossQ.Load()),
		GameRunning: overwatchProcessRunning(),
		GameFocused: isOverwatchForeground(),
		ZoneReady:   t.zoneReady(),
		Zone:        t.getZone(),
		OcrReady:    t.ocrReady(),
		DevMode:     DevMode(),
	}
}

func (t *Tracker) ImportTemplate(kind Outcome, r io.Reader, zone Zone) (TemplateAnalysis, error) {
	img, err := decodeImage(r)
	if err != nil {
		return TemplateAnalysis{}, err
	}
	b := img.Bounds()
	useZone := zone
	if !useZone.Valid() {
		useZone = t.getZone()
	}
	if !useZone.Valid() {
		return TemplateAnalysis{}, ErrZoneRequired
	}
	crop, ok := cropZone(img, useZone)
	if !ok {
		return TemplateAnalysis{}, ErrZoneInvalid
	}
	if err := t.saveZone(useZone); err != nil {
		return TemplateAnalysis{}, err
	}
	h, err := computeTemplateHash(crop)
	if err != nil {
		return TemplateAnalysis{}, err
	}
	hashStr := h
	winT, lossT := t.hashes.get()
	other := ""
	switch kind {
	case OutcomeWin:
		other = lossT
	case OutcomeLoss:
		other = winT
	default:
		return TemplateAnalysis{}, fmt.Errorf("unknown kind")
	}
	analysis := analyzeImport(kind, crop, hashStr, other, b.Dx(), b.Dy(), false)
	if other != "" {
		otherDist := bestDistance(crop, other)
		if kind == OutcomeWin && otherDist >= 0 && otherDist <= hashDistanceThreshold+6 {
			analysis.Warnings = append(analysis.Warnings, "looks_like_loss")
			analysis.OK = false
		}
		if kind == OutcomeLoss && otherDist >= 0 && otherDist <= hashDistanceThreshold+6 {
			analysis.Warnings = append(analysis.Warnings, "looks_like_win")
			analysis.OK = false
		}
	}
	// Hard reject only when win/loss samples would collide at match time.
	if containsWarn(analysis.Warnings, "too_similar_to_other") {
		return analysis, ErrTemplateTooSimilar
	}
	if containsWarn(analysis.Warnings, "looks_like_loss") || containsWarn(analysis.Warnings, "looks_like_win") {
		return analysis, ErrTemplateWrongKind
	}

	saved := normalizeSaved(loadSavedHashes(t.dataDir))
	saved.Zone = useZone
	saved.CustomZone = true
	cb := crop.Bounds()
	switch kind {
	case OutcomeWin:
		saved.Win = hashStr
		saved.CustomWin = true
		saved.WinQuality = analysis.QualityPct
		saved.WinCropW = cb.Dx()
		saved.WinCropH = cb.Dy()
	case OutcomeLoss:
		saved.Loss = hashStr
		saved.CustomLoss = true
		saved.LossQuality = analysis.QualityPct
		saved.LossCropW = cb.Dx()
		saved.LossCropH = cb.Dy()
	}
	if err := saveHashes(t.dataDir, saved); err != nil {
		return analysis, err
	}
	t.applySavedConfig(saved)
	t.reloadTemplateImages()
	_ = saveTemplateCrop(t.dataDir, kind, crop)
	return analysis, nil
}

func containsWarn(list []string, key string) bool {
	for _, w := range list {
		if w == key {
			return true
		}
	}
	return false
}

func (t *Tracker) CaptureTemplate(kind Outcome) (string, error) {
	if !overwatchProcessRunning() {
		return "", ErrGameNotRunning
	}
	x0, y0, bw, bh, ok := overwatchGameBounds()
	if !ok {
		return "", ErrGameNotVisible
	}
	lx, ly, w, h := centerBox(bw, bh)
	img, err := captureRegion(x0+lx, y0+ly, w, h)
	if err != nil {
		return "", err
	}
	hash, err := perceptionHash(img)
	if err != nil {
		return "", err
	}
	s := hash.ToString()
	saved := loadSavedHashes(t.dataDir)
	switch kind {
	case OutcomeWin:
		t.hashes.setWin(s)
		saved.Win = s
	case OutcomeLoss:
		t.hashes.setLoss(s)
		saved.Loss = s
	default:
		return "", fmt.Errorf("unknown kind")
	}
	if err := saveHashes(t.dataDir, saved); err != nil {
		return "", err
	}
	return s, nil
}

func (t *Tracker) SaveZone(zone Zone) error {
	if !zone.Valid() {
		return ErrZoneInvalid
	}
	return t.saveZone(zone)
}

func (t *Tracker) SwapTemplates() error {
	saved := normalizeSaved(loadSavedHashes(t.dataDir))
	if !saved.CustomWin || !saved.CustomLoss {
		return fmt.Errorf("swap requires custom win and loss samples")
	}
	saved.Win, saved.Loss = saved.Loss, saved.Win
	saved.WinQuality, saved.LossQuality = saved.LossQuality, saved.WinQuality
	saved.WinCropW, saved.LossCropW = saved.LossCropW, saved.WinCropW
	saved.WinCropH, saved.LossCropH = saved.LossCropH, saved.WinCropH
	if err := saveHashes(t.dataDir, saved); err != nil {
		return err
	}
	if err := swapTemplateCropFiles(t.dataDir); err != nil {
		return err
	}
	t.applySavedConfig(saved)
	t.reloadTemplateImages()
	return nil
}

func (t *Tracker) ResetToDefaults() error {
	mt := int(t.matchThresholdVal.Load())
	saved := savedHashes{MatchThreshold: mt}
	removeTemplateCrop(t.dataDir, OutcomeWin)
	removeTemplateCrop(t.dataDir, OutcomeLoss)
	if err := saveHashes(t.dataDir, saved); err != nil {
		return err
	}
	t.applySavedConfig(saved)
	t.reloadTemplateImages()
	return nil
}

func (t *Tracker) ClearTemplate(kind Outcome) error {
	saved := normalizeSaved(loadSavedHashes(t.dataDir))
	switch kind {
	case OutcomeWin:
		saved.CustomWin = false
		saved.Win = ""
		saved.WinQuality = 0
		saved.WinCropW = 0
		saved.WinCropH = 0
		removeTemplateCrop(t.dataDir, OutcomeWin)
	case OutcomeLoss:
		saved.CustomLoss = false
		saved.Loss = ""
		saved.LossQuality = 0
		saved.LossCropW = 0
		saved.LossCropH = 0
		removeTemplateCrop(t.dataDir, OutcomeLoss)
	default:
		return fmt.Errorf("unknown kind")
	}
	if saved.CustomWin == false && saved.CustomLoss == false && !saved.CustomZone {
		t.SetEnabled(false)
	}
	if err := saveHashes(t.dataDir, saved); err != nil {
		return err
	}
	t.applySavedConfig(saved)
	return nil
}

func (t *Tracker) Run(ctx context.Context) {
	initDPI()
	log.Printf("owtracker: started (text=%v)", ml.TextReady())
	if note := ml.TextStatusNote(); note != "" {
		log.Printf("owtracker: %s", note)
	}
	for {
		if !sleep(ctx, 0) {
			return
		}

		if !t.enabled.Load() {
			t.state.Store(StateInactive)
			if !sleep(ctx, pollDisabled) {
				return
			}
			continue
		}

		if DevMode() && t.testMode.Load() {
			t.runTestTick(ctx)
			if !sleep(ctx, t.testPollInterval()) {
				return
			}
			continue
		}

		if !overwatchProcessRunning() {
			t.state.Store(StateInactive)
			t.ocr.Close() // free the OCR process until the game starts again
			if !sleep(ctx, pollNoGame) {
				return
			}
			continue
		}

		title := activeWindowTitle()
		t.lastTitle.Store(title)
		foreground := isOverwatchForeground()

		// OBS frames work whether or not the game window has focus; the screen
		// capture fallback inside autoTick needs the game in the foreground.
		if !legacyDetector() {
			if !sleep(ctx, t.autoTick(foreground)) {
				return
			}
			continue
		}

		if !foreground {
			t.state.Store(StateInactive)
			if !sleep(ctx, pollInactive) {
				return
			}
			continue
		}

		st := t.state.Load()
		if st == StateInactive {
			t.state.Store(StateMenu)
			st = StateMenu
		}

		switch st {
		case StateMenu:
			if t.detectMatchHUD() {
				t.state.Store(StateMatch)
				log.Printf("owtracker: STATE_MATCH")
			}
			if t.tryFinish(ctx) {
				continue
			}
			if !sleep(ctx, t.runtimePollInterval()) {
				return
			}
		case StateMatch:
			if t.tryFinish(ctx) {
				continue
			}
			if !sleep(ctx, t.runtimePollInterval()) {
				return
			}
		default:
			if !sleep(ctx, pollInactive) {
				return
			}
		}
	}
}

func (t *Tracker) detectMatchHUD() bool {
	img, err := captureHUD()
	if err != nil {
		return false
	}
	return looksLikeMatchHUD(img)
}

func (t *Tracker) setLastProbeImage(img image.Image) {
	t.lastProbeImgMu.Lock()
	t.lastProbeImg = img
	t.lastProbeImgMu.Unlock()
}

func (t *Tracker) getLastProbeImage() image.Image {
	t.lastProbeImgMu.Lock()
	defer t.lastProbeImgMu.Unlock()
	return t.lastProbeImg
}

func (t *Tracker) runTestTick(ctx context.Context) {
	t.state.Store(StateMenu)
	probe, err := t.ProbeLive()
	if err != nil {
		probe = ProbeResult{
			At:       time.Now(),
			Source:   "live",
			TestMode: true,
			DryRun:   t.dryRun.Load(),
			State:    t.StateName(),
			Notes:    "capture: " + err.Error(),
		}
		t.recordProbe(probe)
		return
	}
	if probe.WouldTrigger && t.dryRun.Load() {
		probe.RejectedReason = "dry_run"
	}
	probe = normalizeTriggerProbe(probe)
	t.recordProbe(probe)

	if t.dryRun.Load() || !probe.WouldTrigger || (probe.MatchMethod != "ocr" && probe.MatchMethod != "banner" && probe.MatchMethod != "ml" && probe.MatchMethod != "text") {
		return
	}
	t.applyIfReady(ctx, probe)
}

func (t *Tracker) applyIfReady(ctx context.Context, probe ProbeResult) bool {
	if t.dryRun.Load() || !probe.WouldTrigger {
		return false
	}
	if !applyProbeAllowed(probe) {
		log.Printf("owtracker: rejected match=%s gold=%d defeat=%d end=%v (color/shape gate)",
			probe.Match, probe.GoldPct, probe.DefeatPct, probe.EndScreenActive)
		return false
	}
	if instantApply(probe) {
		t.applyBannerResult(ctx, probe, "instant")
		return true
	}
	img := t.getLastProbeImage()
	if img == nil {
		return false
	}
	t.bannerStable.add(probe, img)
	if confirmed, ok := t.bannerStable.confirm(time.Now()); ok {
		t.applyBannerResult(ctx, confirmed, "ocr stable")
		return true
	}
	log.Printf("owtracker: candidate match=%s defeat=%d gold=%d (waiting 2nd frame)",
		probe.Match, probe.DefeatPct, probe.GoldPct)
	return false
}

func (t *Tracker) applyBannerResult(ctx context.Context, confirmed ProbeResult, note string) {
	log.Printf("owtracker: confirmed %s method=%s text=%q", confirmed.Match, confirmed.MatchMethod, confirmed.OcrText)
	img := t.getLastProbeImage()
	confirmed.Notes = note
	confirmed.Applied = true
	confirmed.WouldTrigger = true
	t.markApplied(confirmed)
	t.recordProbe(confirmed)
	t.learnFromAuto(confirmed, img)
	if t.onResult != nil {
		t.onResult(Outcome(confirmed.Match))
	}
	t.bannerStable.reset()
	t.endScreen.reset()
	_ = sleep(ctx, cooldown)
}

func (t *Tracker) lastProbeImage() (image.Image, error) {
	if img := t.getLastProbeImage(); img != nil {
		return img, nil
	}
	zone := t.getZone()
	if !zone.Valid() {
		return nil, ErrZoneRequired
	}
	if t.testMode.Load() {
		return t.captureTestZone(zone)
	}
	return captureGameZone(zone)
}

func (t *Tracker) testPollInterval() time.Duration {
	if t.endScreen.isActive() || t.endScreen.isPending() {
		return pollTestFast
	}
	return pollTestSlow
}

func (t *Tracker) runtimePollInterval() time.Duration {
	if t.endScreen.isActive() || t.endScreen.isPending() {
		return pollEndScreen
	}
	if t.state.Load() == StateMatch {
		return pollMatch
	}
	return pollMenu
}

func (t *Tracker) tryFinish(ctx context.Context) bool {
	zone := t.getZone()
	if !zone.Valid() {
		return false
	}
	img, err := captureGameZone(zone)
	if err != nil {
		return false
	}
	hudImg, _ := captureHUD()
	hud := looksLikeMatchHUD(hudImg)
	b := img.Bounds()
	probe := t.probeCrop(img, b.Dx(), b.Dy(), activeWindowTitle(), hud, "runtime")
	t.setLastProbeImage(img)
	if t.debugLogOn.Load() {
		t.recordProbe(probe)
	}
	if !probe.WouldTrigger || (probe.MatchMethod != "ocr" && probe.MatchMethod != "banner" && probe.MatchMethod != "ml" && probe.MatchMethod != "text") {
		return false
	}
	if t.applyIfReady(ctx, probe) {
		t.state.Store(StateMenu)
		return true
	}
	return false
}

func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
