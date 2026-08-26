package owtracker

import (
	"image/color"
	"testing"
	"time"
)

func TestBannerStabilityConfirmOCR(t *testing.T) {
	var buf bannerStability
	now := time.Now()
	img := solid(200, 40, color.RGBA{R: 230, G: 60, B: 60, A: 255})
	base := ProbeResult{
		EndScreenActive: true,
		Match:           "loss",
		MatchMethod:     "ocr",
		OcrText:         "ПОРА",
		WouldTrigger:    true,
		DefeatPct:       35,
		GoldPct:         5,
	}
	for i := 0; i < 3; i++ {
		p := base
		p.At = now.Add(time.Duration(i) * 400 * time.Millisecond)
		buf.add(p, img)
	}
	if _, ok := buf.confirm(now.Add(1200 * time.Millisecond)); !ok {
		t.Fatal("expected stable loss from three agreeing OCR frames")
	}
}

func TestBannerStabilityRejectsBannerColor(t *testing.T) {
	var buf bannerStability
	now := time.Now()
	p := ProbeResult{
		At: now, Match: "loss", MatchMethod: "banner-color",
		EndScreenActive: true,
		GoldPct: 1, DefeatPct: 49, WouldTrigger: true,
	}
	buf.add(p, solid(200, 40, color.RGBA{R: 230, G: 60, B: 60, A: 255}))
	if _, ok := buf.confirm(now); ok {
		t.Fatal("banner-color must not confirm")
	}
}

func TestBannerStabilityRejectsMixed(t *testing.T) {
	var buf bannerStability
	now := time.Now()
	buf.add(ProbeResult{
		At: now, Match: "win", MatchMethod: "ocr", OcrText: "VICTORY", WouldTrigger: true,
		EndScreenActive: true, GoldPct: 40, DefeatPct: 5,
	}, solid(200, 40, color.RGBA{R: 230, G: 180, B: 40, A: 255}))
	buf.add(ProbeResult{
		At: now.Add(300 * time.Millisecond), Match: "loss", MatchMethod: "ocr", OcrText: "DEFEAT", WouldTrigger: true,
		EndScreenActive: true, DefeatPct: 40, GoldPct: 5,
	}, solid(200, 40, color.RGBA{R: 230, G: 60, B: 60, A: 255}))
	if _, ok := buf.confirm(now.Add(400 * time.Millisecond)); ok {
		t.Fatal("expected reject on mixed win/loss")
	}
}

func TestBannerStabilityRejectsWeakColors(t *testing.T) {
	var buf bannerStability
	now := time.Now()
	for i := 0; i < 3; i++ {
		buf.add(ProbeResult{
			At: now.Add(time.Duration(i) * 300 * time.Millisecond),
			Match: "win", MatchMethod: "banner", WouldTrigger: true,
			EndScreenActive: true, GoldPct: 2, DefeatPct: 1,
		}, solid(200, 40, color.RGBA{200, 200, 200, 255}))
	}
	if _, ok := buf.confirm(now.Add(time.Second)); ok {
		t.Fatal("weak colors must not confirm")
	}
}
