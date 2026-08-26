package owtracker

import (
	"image/color"
	"testing"
)

func TestIsBlankCropBlack(t *testing.T) {
	img := solid(100, 50, color.RGBA{A: 255})
	if !isBlankCrop(img) {
		t.Fatal("expected black crop detected")
	}
}

func TestIsBlankCropBannerNotBlank(t *testing.T) {
	img := solid(400, 80, color.RGBA{R: 30, G: 25, B: 25, A: 255})
	for x := 100; x < 300; x++ {
		for y := 25; y < 55; y++ {
			img.Set(x, y, color.RGBA{R: 220, G: 50, B: 50, A: 255})
		}
	}
	if isBlankCrop(img) {
		t.Fatal("defeat banner on dark bg should not be blank")
	}
}

func TestStrongColorDarkBackground(t *testing.T) {
	if strongColorBannerOutcomeForLuma(1, 18, 25) != "loss" {
		t.Fatal("expected lower threshold on dark defeat banner")
	}
}
