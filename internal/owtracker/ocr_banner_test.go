package owtracker

import (
	"image"
	"image/color"
	"testing"
)

func TestFuzzyParseVictoryGarbled(t *testing.T) {
	if parseBannerOutcome("POBEDA") != "win" {
		t.Fatal("expected win for POBEDA")
	}
	if parseBannerOutcome("n POBEDA hen") != "win" {
		t.Fatal("expected win in garbled line")
	}
	if parseBannerOutcome("PORAZHENIE") != "loss" {
		t.Fatal("expected loss for PORAZHENIE")
	}
}

func TestBannerBinaryIgnoresMVPNames(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 100))
	for x := 0; x < 400; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{R: 120, G: 80, B: 200, A: 255}) // purple UI
		}
	}
	if bannerBinaryScore(bannerBinary(img)) >= 120 {
		t.Fatal("purple MVP UI should not look like banner text")
	}
}

func TestBannerBinaryDetectsGoldText(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 80))
	for x := 120; x < 280; x++ {
		for y := 25; y < 55; y++ {
			img.Set(x, y, color.RGBA{R: 230, G: 180, B: 40, A: 255})
		}
	}
	if bannerBinaryScore(bannerBinary(img)) < 40 {
		t.Fatal("expected gold banner blob")
	}
	if !endScreenBannerImage(img) {
		t.Fatal("gold banner crop should pass end screen image check")
	}
	band, ok := extractBannerTextBand(img, 38, 1)
	if !ok || band == nil {
		t.Fatal("expected banner band extraction")
	}
}
