package owtracker

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync/atomic"
	"time"
)

const (
	StateInactive int32 = 0
	StateMenu     int32 = 1
	StateMatch    int32 = 2
)

const (
	pollInactive = 2 * time.Second
	pollMenu     = 1 * time.Second
	pollMatch    = 1 * time.Second
	antiBlink    = 1 * time.Second
	cooldown     = 40 * time.Second
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
	WinQuality  int    `json:"winQuality,omitempty"`
	LossQuality int    `json:"lossQuality,omitempty"`
	GameRunning bool   `json:"gameRunning"`
	GameFocused bool   `json:"gameFocused"`
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
}

func New(dataDir string, onResult func(Outcome)) *Tracker {
	t := &Tracker{dataDir: dataDir, onResult: onResult}
	saved := loadSavedHashes(dataDir)
	win, loss := WIN_HASH_TEMPLATE, LOSS_HASH_TEMPLATE
	if saved.Win != "" {
		win = saved.Win
	}
	if saved.Loss != "" {
		loss = saved.Loss
	}
	t.hashes.setWin(win)
	t.hashes.setLoss(loss)
	t.winQ.Store(int32(saved.WinQuality))
	t.lossQ.Store(int32(saved.LossQuality))
	t.lastTitle.Store("")
	return t
}

func (t *Tracker) Enabled() bool { return t.enabled.Load() }

func (t *Tracker) SetEnabled(v bool) { t.enabled.Store(v) }

func (t *Tracker) Toggle() bool {
	for {
		cur := t.enabled.Load()
		if t.enabled.CompareAndSwap(cur, !cur) {
			return !cur
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
	title := activeWindowTitle()
	t.lastTitle.Store(title)
	return Status{
		Enabled:     t.Enabled(),
		State:       t.StateName(),
		WindowTitle: title,
		WinReady:    win != "",
		LossReady:   loss != "",
		WinQuality:  int(t.winQ.Load()),
		LossQuality: int(t.lossQ.Load()),
		GameRunning: overwatchProcessRunning(),
		GameFocused: isOverwatchTitle(title),
	}
}

func (t *Tracker) ImportTemplate(kind Outcome, r io.Reader) (TemplateAnalysis, error) {
	img, err := decodeImage(r)
	if err != nil {
		return TemplateAnalysis{}, err
	}
	b := img.Bounds()
	crop, usedFull := templateRegion(img)
	h, err := perceptionHash(crop)
	if err != nil {
		return TemplateAnalysis{}, err
	}
	hashStr := h.ToString()
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
	analysis := analyzeImport(kind, crop, hashStr, other, b.Dx(), b.Dy(), usedFull)
	// Hard reject only when win/loss samples would collide at match time.
	if containsWarn(analysis.Warnings, "too_similar_to_other") {
		return analysis, ErrTemplateTooSimilar
	}

	saved := loadSavedHashes(t.dataDir)
	switch kind {
	case OutcomeWin:
		t.hashes.setWin(hashStr)
		t.winQ.Store(int32(analysis.QualityPct))
		saved.Win = hashStr
		saved.WinQuality = analysis.QualityPct
	case OutcomeLoss:
		t.hashes.setLoss(hashStr)
		t.lossQ.Store(int32(analysis.QualityPct))
		saved.Loss = hashStr
		saved.LossQuality = analysis.QualityPct
	}
	if err := saveHashes(t.dataDir, saved); err != nil {
		return analysis, err
	}
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

func (t *Tracker) ClearTemplate(kind Outcome) error {
	saved := loadSavedHashes(t.dataDir)
	switch kind {
	case OutcomeWin:
		t.hashes.setWin("")
		t.winQ.Store(0)
		saved.Win = ""
		saved.WinQuality = 0
	case OutcomeLoss:
		t.hashes.setLoss("")
		t.lossQ.Store(0)
		saved.Loss = ""
		saved.LossQuality = 0
	default:
		return fmt.Errorf("unknown kind")
	}
	if saved.Win == "" || saved.Loss == "" {
		t.SetEnabled(false)
	}
	return saveHashes(t.dataDir, saved)
}

func (t *Tracker) Run(ctx context.Context) {
	initDPI()
	log.Printf("owtracker: started (win/loss pHash + state machine)")
	for {
		if !sleep(ctx, 0) {
			return
		}
		if !t.enabled.Load() {
			t.state.Store(StateInactive)
			if !sleep(ctx, pollInactive) {
				return
			}
			continue
		}

		title := activeWindowTitle()
		t.lastTitle.Store(title)
		if !isOverwatchTitle(title) {
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
			if !sleep(ctx, pollMenu) {
				return
			}
		case StateMatch:
			if t.tryFinish(ctx) {
				continue
			}
			if !sleep(ctx, pollMatch) {
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

func (t *Tracker) tryFinish(ctx context.Context) bool {
	kind, dist, ok := t.classifyCenter()
	if !ok {
		return false
	}
	if !sleep(ctx, antiBlink) {
		return true
	}
	kind2, dist2, ok2 := t.classifyCenter()
	if !ok2 || kind2 != kind {
		log.Printf("owtracker: anti-blink rejected (%s d=%d then %s d=%d)", kind, dist, kind2, dist2)
		return false
	}
	log.Printf("owtracker: confirmed %s (hamming %d then %d)", kind, dist, dist2)
	if t.onResult != nil {
		t.onResult(kind)
	}
	t.state.Store(StateMenu)
	_ = sleep(ctx, cooldown)
	return true
}

func (t *Tracker) classifyCenter() (Outcome, int, bool) {
	img, err := captureCenter()
	if err != nil {
		return "", -1, false
	}
	h, err := perceptionHash(img)
	if err != nil {
		return "", -1, false
	}
	winT, lossT := t.hashes.get()
	if d, ok := hashMatches(h, winT); ok {
		return OutcomeWin, d, true
	}
	if d, ok := hashMatches(h, lossT); ok {
		return OutcomeLoss, d, true
	}
	return "", -1, false
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
