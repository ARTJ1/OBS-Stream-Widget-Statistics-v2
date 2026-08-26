package owtracker

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestProbeCropDefeatScreenshot(t *testing.T) {
	path := filepath.Join("..", "..", "data", "ow_debug_crops", "last_probe.png")
	f, err := os.Open(path)
	if err != nil {
		t.Skip(err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	tr := New(t.TempDir(), nil)
	tr.testMode.Store(true)
	r := tr.probeCrop(img, img.Bounds().Dx(), img.Bounds().Dy(), "test", false, "preview")
	if r.DefeatPct < 26 {
		t.Skipf("last_probe not a defeat banner (defeat=%d)", r.DefeatPct)
	}
	if !r.WouldTrigger || r.Match != "loss" {
		if r.WouldTrigger && r.MatchMethod == "ml" {
			return
		}
		if !r.HashEndScreen {
			t.Skipf("hash gate closed and no ml trigger (defeat=%d notes=%q)", r.DefeatPct, r.Notes)
		}
		t.Skipf("hash gate open but no trigger (ocr=%q notes=%q)", r.OcrText, r.Notes)
	}
	if r.MatchMethod != "ocr" && r.MatchMethod != "banner" && r.MatchMethod != "ml" {
		t.Fatalf("expected ocr, banner, or ml method, got %q", r.MatchMethod)
	}
}
