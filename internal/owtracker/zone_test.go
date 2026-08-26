package owtracker

import (
	"image"
	"image/color"
	"testing"
)

func TestZonePixelBox(t *testing.T) {
	z := Zone{XPct: 10, YPct: 20, WPct: 30, HPct: 10}
	x, y, w, h := z.PixelBox(1920, 1080)
	if x != 192 || y != 216 || w != 576 || h != 108 {
		t.Fatalf("got %d,%d %dx%d", x, y, w, h)
	}
}

func TestCropZone(t *testing.T) {
	full := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			full.Set(x, y, color.RGBA{R: 10, G: 10, B: 10, A: 255})
		}
	}
	z := Zone{XPct: 25, YPct: 25, WPct: 50, HPct: 50}
	crop, ok := cropZone(full, z)
	if !ok {
		t.Fatal("expected crop")
	}
	if crop.Bounds().Dx() != 100 || crop.Bounds().Dy() != 50 {
		t.Fatalf("crop size %dx%d", crop.Bounds().Dx(), crop.Bounds().Dy())
	}
}
