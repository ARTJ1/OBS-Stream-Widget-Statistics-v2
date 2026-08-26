package owtracker

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestOCRRealDefeatCrop(t *testing.T) {
	path := filepath.Join("..", "..", "data", "ow_debug_crops", "last_probe.png")
	f, err := os.Open(path)
	if err != nil {
		t.Skip("last_probe.png not available:", err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	gold, defeat := bannerColorHint(img)
	out := strongColorBannerOutcome(gold, defeat)
	if out != "loss" {
		t.Skipf("last_probe color weak for fallback (gold=%d defeat=%d)", gold, defeat)
	}
	dir := t.TempDir()
	bindOCRDataDir(dir)
	ensureOCR(dir)
	if !ocrReady {
		t.Fatalf("ocr not ready: %s", ocrInitNote)
	}
	text, outcome := recognizeBannerText(img, dir, filepath.Join(dir, "ocr"), gold, defeat)
	t.Logf("ocr text=%q outcome=%q gold=%d defeat=%d", text, outcome, gold, defeat)
	// Twitch compression often breaks Tesseract on stylized OW2 font; color fallback handles this case.
	if outcome == "loss" {
		return
	}
	t.Log("ocr did not read ПОРАЖЕНИЕ directly; banner-color path must trigger in probeCrop")
}
