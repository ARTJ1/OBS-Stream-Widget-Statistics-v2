package owtracker

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDebugDefeatOCR(t *testing.T) {
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
	gold, defeat := bannerColorHint(img)
	t.Logf("gold=%d defeat=%d color=%q", gold, defeat, strongColorBannerOutcome(gold, defeat))
	variants := prepareOCRVariants(img, gold, defeat)
	t.Logf("variants=%d", len(variants))
	dir := t.TempDir()
	bindOCRDataDir(dir)
	ensureOCR(dir)
	if !ocrReady {
		t.Fatal(ocrInitNote)
	}
	for i, v := range variants {
		p := filepath.Join(dir, fmt.Sprintf("v%d.png", i))
		if err := writeOCRPNG(p, v); err != nil {
			t.Fatal(err)
		}
		for _, psm := range []string{"8", "7", "13"} {
			text, err := runTesseractPSM(p, psm)
			t.Logf("v%d psm=%s err=%v text=%q out=%q", i, psm, err, text, parseBannerOutcome(text))
		}
	}
}
