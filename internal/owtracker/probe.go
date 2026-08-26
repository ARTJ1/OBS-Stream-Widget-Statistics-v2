package owtracker

import (
	"context"
	"image"
	"io"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"
)

func (t *Tracker) DebugStatus() DebugStatus {
	aw, al := t.autoLearn.counts()
	return DebugStatus{
		TestMode:       t.testMode.Load(),
		DebugLog:       t.debugLogOn.Load(),
		DryRun:         t.dryRun.Load(),
		CaptureSource:  t.getCaptureSource(),
		MatchThreshold: t.matchThreshold(),
		OcrReady:       t.ocrReady(),
		MLReady:        ml.Ready(),
		MLNote:         ml.StatusNote(),
		TextReady:      ml.TextReady(),
		TextNote:       ml.TextStatusNote(),
		MLAutoWin:      aw,
		MLAutoLoss:     al,
		LogPath:        t.debug.path,
		EntryCount:     t.debug.count(),
		LastProbe:      t.debug.last(),
		LastApplied:    t.getLastApplied(),
	}
}

func (t *Tracker) getLastApplied() *ProbeResult {
	t.lastAppliedMu.Lock()
	defer t.lastAppliedMu.Unlock()
	if t.lastApplied == nil {
		return nil
	}
	cp := *t.lastApplied
	return &cp
}

func (t *Tracker) markApplied(r ProbeResult) {
	cp := r
	cp.Applied = true
	cp.WouldTrigger = true
	t.lastAppliedMu.Lock()
	t.lastApplied = &cp
	t.lastAppliedMu.Unlock()
}

func (t *Tracker) SetTestMode(v bool)  { t.testMode.Store(v); t.persistDebugSettings() }

// ApplyIfReady scores win/loss when probe detection already passed (test/manual probe).
func (t *Tracker) ApplyIfReady(ctx context.Context, probe ProbeResult) bool {
	probe = normalizeTriggerProbe(probe)
	return t.applyIfReady(ctx, probe)
}
func (t *Tracker) SetDebugLog(v bool)  { t.debugLogOn.Store(v); t.persistDebugSettings() }
func (t *Tracker) SetDryRun(v bool)    { t.dryRun.Store(v); t.persistDebugSettings() }

func (t *Tracker) RecentProbes(limit int) []ProbeResult {
	return t.debug.recent(limit)
}

func (t *Tracker) ClearDebugLog() error {
	return t.debug.clear()
}

// ProbeImage classifies an uploaded screenshot using the configured zone.
func (t *Tracker) ProbeImage(img image.Image, source string) ProbeResult {
	b := img.Bounds()
	zone := t.getZone()
	crop, ok := cropZone(img, zone)
	if !ok {
		return ProbeResult{Source: source, Notes: "zone not configured"}
	}
	return t.probeCrop(crop, b.Dx(), b.Dy(), "", false, source)
}

// ProbeLive captures the configured zone from screen/window or Overwatch.
func (t *Tracker) ProbeLive() (ProbeResult, error) {
	if t.testMode.Load() {
		return t.probeLiveTest()
	}
	zone := t.getZone()
	if !zone.Valid() {
		return ProbeResult{}, ErrZoneRequired
	}
	img, err := captureGameZone(zone)
	if err != nil {
		return ProbeResult{}, err
	}
	b := img.Bounds()
	hud := false
	if t.testMode.Load() {
		if hudImg, err := captureHUDTest(); err == nil {
			hud = looksLikeMatchHUD(hudImg)
		}
	} else if hudImg, err := captureHUD(); err == nil {
		hud = looksLikeMatchHUD(hudImg)
	}
	r := t.probeCrop(img, b.Dx(), b.Dy(), activeWindowTitle(), hud, "live")
	r.CropSaved = saveDebugCrop(t.dataDir, "last_probe.png", img)
	return r, nil
}

func probeBestScore(r ProbeResult) int {
	if r.WinPixelPct >= 0 && r.LossPixelPct >= 0 {
		return 100 - absInt(r.WinPixelPct-r.LossPixelPct)
	}
	score := 999
	if r.WinDistance >= 0 && r.WinDistance < score {
		score = r.WinDistance
	}
	if r.LossDistance >= 0 && r.LossDistance < score {
		score = r.LossDistance
	}
	return score
}

func (t *Tracker) recordProbe(r ProbeResult) {
	if !t.debugLogOn.Load() && !t.testMode.Load() {
		return
	}
	_ = t.debug.append(r)
}

// RecordProbePublic always saves a probe entry (manual test / upload).
func (t *Tracker) RecordProbePublic(r ProbeResult) {
	_ = t.debug.append(r)
}

// DecodeImagePublic decodes an uploaded screenshot for probe tests.
func DecodeImagePublic(r io.Reader) (image.Image, error) {
	return decodeImage(r)
}
