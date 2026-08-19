package owtracker

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestCenterBox(t *testing.T) {
	x, y, w, h := centerBox(1920, 1080)
	if w != 480 || h != 108 {
		t.Fatalf("size got %dx%d", w, h)
	}
	if x != 720 || y != 486 {
		t.Fatalf("origin got %d,%d", x, y)
	}
}

func TestIsOverwatchTitle(t *testing.T) {
	if !isOverwatchTitle("Overwatch") || !isOverwatchTitle(" Overwatch 2 ") {
		t.Fatal("expected Overwatch titles to match")
	}
	if isOverwatchTitle("Chrome") || isOverwatchTitle("Overwatch 2 - Battle.net") {
		t.Fatal("expected non-game titles to be rejected")
	}
}

func TestHashDistanceSameImage(t *testing.T) {
	img := solid(80, 40, color.RGBA{R: 220, G: 180, B: 40, A: 255})
	h1, err := perceptionHash(img)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := perceptionHash(img)
	if err != nil {
		t.Fatal(err)
	}
	d, err := h1.Distance(h2)
	if err != nil || d != 0 {
		t.Fatalf("distance %d err %v", d, err)
	}
	if dist, ok := hashMatches(h1, h2.ToString()); !ok || dist != 0 {
		t.Fatalf("template match failed dist=%d ok=%v", dist, ok)
	}
}

func TestLooksLikeMatchHUD(t *testing.T) {
	dark := solid(40, 40, color.RGBA{R: 10, G: 10, B: 12, A: 255})
	if looksLikeMatchHUD(dark) {
		t.Fatal("dark crop should not look like HUD")
	}
	hud := image.NewRGBA(image.Rect(0, 0, 40, 40))
	draw.Draw(hud, hud.Bounds(), &image.Uniform{C: color.RGBA{R: 8, G: 10, B: 14, A: 255}}, image.Point{}, draw.Src)
	for y := 8; y < 20; y++ {
		for x := 10; x < 30; x++ {
			hud.Set(x, y, color.RGBA{R: 240, G: 240, B: 245, A: 255})
		}
	}
	for y := 22; y < 30; y++ {
		for x := 6; x < 14; x++ {
			hud.Set(x, y, color.RGBA{R: 70, G: 140, B: 230, A: 255})
		}
		for x := 26; x < 34; x++ {
			hud.Set(x, y, color.RGBA{R: 230, G: 90, B: 40, A: 255})
		}
	}
	if !looksLikeMatchHUD(hud) {
		t.Fatal("synthetic HUD should match")
	}
}

func solid(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	return img
}
