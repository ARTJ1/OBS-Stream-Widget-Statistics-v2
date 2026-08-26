package owtracker

import (
	"image/color"
	"testing"
)

func TestPixelSimilaritySameImage(t *testing.T) {
	img := solid(480, 180, color.RGBA{R: 200, G: 160, B: 30, A: 255})
	score := pixelSimilarityPct(img, img)
	if score < 95 {
		t.Fatalf("expected ~100 self similarity, got %d", score)
	}
}

func TestPickMatchScoresBlocksFalseLossOnGold(t *testing.T) {
	// User case: loss pixel 56% vs win 44%, but gold victory banner visible.
	img := solid(480, 180, color.RGBA{R: 220, G: 170, B: 50, A: 255})
	match, ok, _, method := pickMatchScores(36, 28, 44, 56, 8, img)
	if ok && match == "loss" {
		t.Fatalf("expected no loss on gold banner, got match=%q ok=%v method=%s", match, ok, method)
	}
	if !ok || match != "win" || method != "color-victory" {
		t.Fatalf("expected color-victory win, got match=%q ok=%v method=%s", match, ok, method)
	}
}

func TestPickMatchScoresPixelWin(t *testing.T) {
	img := solid(480, 180, color.RGBA{R: 30, G: 30, B: 30, A: 255})
	match, ok, _, method := pickMatchScores(30, 32, 78, 42, 8, img)
	if !ok || match != "win" || method != "pixel" {
		t.Fatalf("expected pixel win, got match=%q ok=%v method=%s", match, ok, method)
	}
}

func TestPickMatchScoresHashRelativeNeedsColors(t *testing.T) {
	img := solid(480, 180, color.RGBA{R: 30, G: 30, B: 30, A: 255})
	match, ok, _, _ := pickMatchScores(22, 35, 20, 18, 8, img)
	if ok {
		t.Fatalf("hash-only match without banner colors must not trigger, got %q", match)
	}
}

func TestColorBlocksLoss(t *testing.T) {
	if !colorBlocksLoss(15, 3) {
		t.Fatal("expected gold banner to block loss")
	}
}
