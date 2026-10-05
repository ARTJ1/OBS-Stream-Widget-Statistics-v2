package owtracker

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadSample(t *testing.T, path string) image.Image {
	t.Helper()
	fh, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	img, err := decodeImage(fh)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

// The cheap stage-1 strip must fire on every real banner and never on gameplay.
func TestStripSeparatesBanners(t *testing.T) {
	files, _ := filepath.Glob("testdata/banners/*.png")
	for _, f := range files {
		if strings.Contains(f, "flash") {
			continue // fly-in flash: only the banner CNN is expected to read it
		}
		g, r := stripColorPct(stripFromImage(loadSample(t, f)))
		want := !strings.HasPrefix(filepath.Base(f), "none")
		if got := stripLooksLikeBanner(g, r); got != want {
			t.Errorf("%s: strip hit=%v want %v (gold=%d%% red=%d%%)", filepath.Base(f), got, want, g, r)
		}
	}
}

func loadSampleB(b *testing.B, path string) image.Image {
	fh, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer fh.Close()
	img, err := decodeImage(fh)
	if err != nil {
		b.Fatal(err)
	}
	return img
}
