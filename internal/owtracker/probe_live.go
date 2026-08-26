package owtracker

import (
	"fmt"
	"image"
	"strconv"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"
)

func (t *Tracker) CapturePreview() (image.Image, error) {
	zone := t.getZone()
	if !zone.Valid() {
		return nil, ErrZoneRequired
	}
	// Admin preview: capture from screen/window so it works without Overwatch running.
	return t.captureTestZone(zone)
}

func (t *Tracker) PreviewProbe() (ProbeResult, error) {
	zone := t.getZone()
	if !zone.Valid() {
		return ProbeResult{}, ErrZoneRequired
	}
	img, err := t.captureTestZone(zone)
	if err != nil {
		return ProbeResult{}, err
	}
	b := img.Bounds()
	return t.probeCrop(img, b.Dx(), b.Dy(), activeWindowTitle(), false, "preview"), nil
}

func (t *Tracker) captureTestZone(zone Zone) (image.Image, error) {
	img, _, err := t.captureTestZoneWithTitle(zone)
	return img, err
}

func (t *Tracker) captureTestZoneWithTitle(zone Zone) (image.Image, string, error) {
	windowFn := captureWindowZone
	if t.testMode.Load() {
		windowFn = captureWindowZoneSmart
	}
	switch t.getCaptureSource() {
	case CaptureWindow:
		return windowFn(zone)
	case CaptureScreen:
		return captureScreenZone(zone)
	default:
		// In test mode screen capture picks the wrong monitor region on Twitch; window only.
		if t.testMode.Load() {
			return windowFn(zone)
		}
		img, title, err := windowFn(zone)
		if err == nil {
			return img, title, nil
		}
		return captureScreenZone(zone)
	}
}

func (t *Tracker) probeLiveTest() (ProbeResult, error) {
	zone := t.getZone()
	if !zone.Valid() {
		return ProbeResult{}, ErrZoneRequired
	}

	type attempt struct {
		label string
		fn    func(Zone) (image.Image, string, error)
	}
	windowFn := captureWindowZone
	if t.testMode.Load() {
		windowFn = captureWindowZoneSmart
	}
	var tries []attempt
	switch t.getCaptureSource() {
	case CaptureWindow:
		tries = []attempt{{CaptureWindow, windowFn}}
	case CaptureScreen:
		tries = []attempt{{CaptureScreen, captureScreenZone}}
	default:
		if t.testMode.Load() {
			tries = []attempt{{CaptureWindow, windowFn}}
		} else {
			tries = []attempt{
				{CaptureWindow, windowFn},
				{CaptureScreen, captureScreenZone},
			}
		}
	}

	var best ProbeResult
	var bestImg image.Image
	bestScore := 999999
	var lastErr error
	for _, a := range tries {
		img, title, err := a.fn(zone)
		if err != nil {
			lastErr = err
			continue
		}
		hud := false
		if !t.testMode.Load() {
			if hudImg, err := captureHUDTest(); err == nil {
				hud = looksLikeMatchHUD(hudImg)
			}
		}
		b := img.Bounds()
		r := t.probeCrop(img, b.Dx(), b.Dy(), title, hud, "live:"+a.label)
		if isBlankCrop(img) {
			lastErr = ErrBlankCapture
			continue
		}
		score := probeLivePickScore(r)
		if score < bestScore {
			best = r
			bestImg = img
			bestScore = score
		}
	}
	if bestScore == 999999 {
		if lastErr != nil {
			return ProbeResult{}, lastErr
		}
		return ProbeResult{}, ErrGameNotVisible
	}
	if bestImg != nil {
		best.CropSaved = saveDebugCrop(t.dataDir, "last_probe.png", bestImg)
		t.setLastProbeImage(bestImg)
	}
	return best, nil
}

// probeLivePickScore ranks live captures: ML/OCR trigger beats hash-only frames.
func probeLivePickScore(r ProbeResult) int {
	if r.WouldTrigger && (r.MatchMethod == "ml" || r.MatchMethod == "ocr") {
		return 0
	}
	if r.WouldTrigger && r.MatchMethod == "banner" {
		return 10
	}
	if r.HashEndScreen || r.EndScreenActive {
		return 50
	}
	if r.OcrText != "" {
		return 80
	}
	return 1000 + probeBestScore(r)
}

func (t *Tracker) probeCrop(crop image.Image, sourceW, sourceH int, windowTitle string, hud bool, source string) ProbeResult {
	b := crop.Bounds()
	r := ProbeResult{
		At:          time.Now(),
		Source:      source,
		WindowTitle: windowTitle,
		CropW:       b.Dx(),
		CropH:       b.Dy(),
		Threshold:   t.matchThreshold(),
		State:       t.StateName(),
		TestMode:    t.testMode.Load(),
		DryRun:      t.dryRun.Load(),
		HUDDetected: hud,
	}
	if isBlankCrop(crop) {
		r.AvgLuminance = cropAvgLuminance(crop)
		r.BlankCrop = true
		r.Notes = "захват пустой (чёрный кадр) — проверьте окно Opera/Twitch"
		return r
	}
	if b.Dx() < 400 || b.Dy() < 80 {
		r.Notes = "захват слишком маленький — откройте Twitch/Overwatch на весь экран"
		return r
	}
	if sourceW > 0 && sourceH > 0 && (sourceW != b.Dx() || sourceH != b.Dy()) {
		r.Notes = "cropped from " + strconv.Itoa(sourceW) + "x" + strconv.Itoa(sourceH)
	}

	winT, lossT := t.hashes.get()
	if winT == "" && lossT == "" {
		r.Notes = "templates missing"
		return r
	}

	winW, winH, lossW, lossH := t.templateCropSizes()
	_ = winW
	_ = winH
	_ = lossW
	_ = lossH
	winDist, lossDist, winPix, lossPix := t.scoreCrop(crop, winT, lossT)
	r.WinDistance = winDist
	r.LossDistance = lossDist
	r.WinPixelPct = winPix
	r.LossPixelPct = lossPix
	r.GoldPct, r.DefeatPct = bannerColorHint(crop)
	r.AvgLuminance = cropAvgLuminance(crop)
	if winDist >= 0 {
		r.WinSimilarity = similarityPct(winDist)
	}
	if lossDist >= 0 {
		r.LossSimilarity = similarityPct(lossDist)
	}

	threshold := t.matchThreshold()
	hashEnd := isHashEndScreen(winDist, lossDist, winPix, lossPix, threshold)
	r.HashEndScreen = hashEnd
	t.endScreen.tick(hud, crop, t.testMode.Load(), hashEnd)
	endActive := t.endScreen.isActive()
	r.EndScreenActive = endActive
	endGate := applyEndScreenGate(endActive, source)
	mlApplyOK := b.Dx() >= 600 && b.Dy() >= 150

	// v6: binary banner TEXT net (white glyphs on black) — not pHash/color matching.
	if ml.TextReady() && endGate && mlApplyOK {
		tr, _ := ml.ClassifyBannerText(crop, r.GoldPct, r.DefeatPct)
		r.MLWinPct = int(tr.WinProb * 100)
		r.MLLossPct = int(tr.LossProb * 100)
		r.MLNonePct = int(tr.NoneProb * 100)
		if tr.Label == ml.LabelWin {
			r.MLConfidence = r.MLWinPct
		} else if tr.Label == ml.LabelLoss {
			r.MLConfidence = r.MLLossPct
		} else {
			r.MLConfidence = r.MLNonePct
		}
		if label, ok := textAccepts(tr); ok {
			r.Match = label
			r.WouldTrigger = true
			r.MatchMethod = "text"
			r.Notes = fmt.Sprintf("text: %s (win=%d%% loss=%d%% none=%d%%)", label, r.MLWinPct, r.MLLossPct, r.MLNonePct)
			return r
		}
		if tr.Label != ml.LabelNone {
			r.Notes = fmt.Sprintf("text слабый: %s (win=%d loss=%d)", tr.Label, r.MLWinPct, r.MLLossPct)
		} else {
			r.Notes = fmt.Sprintf("text: none (win=%d%% loss=%d%% none=%d%%)", r.MLWinPct, r.MLLossPct, r.MLNonePct)
		}
	}

	// Hash/color metrics kept for debug display only — never used for score trigger in v6.
	if !endActive && !endGate {
		r.Notes = "ожидание баннера (меню/матч)"
	} else if ml.TextReady() && r.Notes == "" {
		r.Notes = "text: баннер не распознан"
	} else if !ml.TextReady() && r.Notes == "" {
		r.Notes = "text net не загружена"
	}

	// Legacy metrics for debug display only (no score trigger).
	_, wouldTrigger, ambiguous, method := pickMatchScores(winDist, lossDist, winPix, lossPix, threshold, crop)
	_ = wouldTrigger
	_ = method
	if ambiguous && r.Notes == "" {
		r.Ambiguous = true
		r.Notes = "hash/pixel ambiguous (ocr mode)"
	}
	return r
}
