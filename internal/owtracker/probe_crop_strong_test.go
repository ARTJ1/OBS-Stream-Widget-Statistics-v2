package owtracker

import (
	"image/color"
	"testing"
)

func TestProbeLivePickScorePrefersOCRTrigger(t *testing.T) {
	trigger := ProbeResult{WouldTrigger: true, MatchMethod: "ocr", Match: "loss"}
	weak := ProbeResult{WinDistance: 10, LossDistance: 40, HashEndScreen: true}
	if probeLivePickScore(trigger) >= probeLivePickScore(weak) {
		t.Fatal("OCR trigger frame must outrank hash-only frame")
	}
}

func TestProbeCropStrongDefeatNoMLFalsePositive(t *testing.T) {
	tr := New(t.TempDir(), nil)
	tr.testMode.Store(true)
	img := solid(400, 120, color.RGBA{R: 30, G: 25, B: 25, A: 255})
	for x := 0; x < 400; x++ {
		for y := 30; y < 90; y++ {
			img.Set(x, y, color.RGBA{R: 220, G: 40, B: 40, A: 255})
		}
	}
	r := tr.probeCrop(img, 400, 120, "test", false, "upload")
	if r.WouldTrigger {
		t.Fatalf("color-only crop must not trigger score (ML v5), got match=%q notes=%q", r.Match, r.Notes)
	}
}
