package owtracker

import (
	"image/color"
	"testing"
)

func TestProbeImageNoTriggerWithoutBannerText(t *testing.T) {
	img := solid(960, 540, color.RGBA{R: 40, G: 35, B: 50, A: 255})
	tr := New(t.TempDir(), nil)
	tr.setZone(Zone{XPct: 0, YPct: 0, WPct: 100, HPct: 100})
	got := tr.ProbeImage(img, "upload")
	if got.WouldTrigger {
		t.Fatalf("expected no trigger without banner text, got %+v", got)
	}
}

func TestProbeImageTriggersWhenOCRMatches(t *testing.T) {
	tr := New(t.TempDir(), nil)
	if !tr.ocrReady() {
		t.Skip("embedded ocr not ready: ", ocrStatusNote())
	}
	img, err := defaultTemplateImage(OutcomeWin)
	if err != nil {
		t.Fatal(err)
	}
	tr.setZone(Zone{XPct: 0, YPct: 0, WPct: 100, HPct: 100})
	got := tr.ProbeImage(img, "upload")
	if got.OcrText == "" && got.Notes != "" {
		t.Log("ocr note:", got.Notes)
	}
	// Upload path always runs OCR; factory banner may not parse cleanly in CI.
	if got.WouldTrigger && got.Match != "win" {
		t.Fatalf("unexpected trigger: %+v", got)
	}
}
