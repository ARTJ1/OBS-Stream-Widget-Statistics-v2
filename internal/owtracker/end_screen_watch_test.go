package owtracker

import (
	"image/color"
	"testing"
)

func TestEndScreenBannerStrict(t *testing.T) {
	if endScreenBanner(8, 3) {
		t.Fatal("weak gold should not be end screen")
	}
	if endScreenBanner(15, 5) {
		t.Fatal("15% gold should not be end screen")
	}
	if !endScreenBanner(25, 8) {
		t.Fatal("expected strong victory banner")
	}
	if !endScreenBanner(8, 25) {
		t.Fatal("expected strong defeat banner")
	}
	if endScreenBanner(20, 18) {
		t.Fatal("similar gold/defeat mix should not be end screen")
	}
}

func TestEndScreenWatchNeedsTwoFrames(t *testing.T) {
	var w endScreenWatch
	weak := solid(480, 180, color.RGBA{R: 140, G: 110, B: 90, A: 255})
	w.tick(false, weak, true, false)
	if w.isActive() {
		t.Fatal("weak banner should not activate on one frame")
	}
	w.tick(false, weak, true, false)
	if !w.isActive() {
		t.Fatal("two weak banner frames should activate end screen")
	}
}

func TestEndScreenWatchStrongBannerFast(t *testing.T) {
	var w endScreenWatch
	img := solid(480, 180, color.RGBA{R: 220, G: 170, B: 50, A: 255})
	w.tick(false, img, true, false)
	if !w.isActive() {
		t.Fatal("strong gold banner should activate immediately")
	}
}

func TestEndScreenWatchHashGateFast(t *testing.T) {
	var w endScreenWatch
	dark := solid(480, 180, color.RGBA{R: 30, G: 25, B: 40, A: 255})
	w.tick(false, dark, true, true)
	if !w.isActive() {
		t.Fatal("hash end-screen signal should activate OCR window immediately")
	}
}

func TestEndScreenWatchGameplayDark(t *testing.T) {
	var w endScreenWatch
	dark := solid(480, 180, color.RGBA{R: 30, G: 25, B: 40, A: 255})
	for i := 0; i < 5; i++ {
		w.tick(false, dark, true, false)
	}
	if w.isActive() {
		t.Fatal("dark gameplay should not activate end screen")
	}
}
