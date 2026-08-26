package owtracker

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestTemplateRegionFullScreenshot(t *testing.T) {
	img := solidRGBA(1920, 1080, color.RGBA{R: 30, G: 40, B: 50, A: 255})
	// Paint a bright banner in the center crop zone.
	x, y, w, h := centerBox(1920, 1080)
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			img.Set(xx, yy, color.RGBA{R: 220, G: 180, B: 40, A: 255})
		}
	}
	crop, usedFull := templateRegion(img)
	if usedFull {
		t.Fatal("expected center crop for full screenshot")
	}
	if crop.Bounds().Dx() != w || crop.Bounds().Dy() != h {
		t.Fatalf("crop size %dx%d want %dx%d", crop.Bounds().Dx(), crop.Bounds().Dy(), w, h)
	}
}

func TestTemplateRegionSmallBanner(t *testing.T) {
	img := solidRGBA(480, 108, color.RGBA{R: 200, G: 160, B: 30, A: 255})
	crop, usedFull := templateRegion(img)
	if !usedFull {
		t.Fatal("expected full image for small banner crop")
	}
	if crop.Bounds().Dx() != 480 {
		t.Fatalf("got %d", crop.Bounds().Dx())
	}
}

func TestImportTemplateFromPNG(t *testing.T) {
	dir := t.TempDir()
	tr := New(dir, nil)
	img := solidRGBA(1920, 1080, color.RGBA{R: 20, G: 20, B: 30, A: 255})
	x, y, w, h := centerBox(1920, 1080)
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			img.Set(xx, yy, color.RGBA{R: 240, G: 200, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	a, err := tr.ImportTemplate(OutcomeWin, &buf, DefaultZone())
	if err != nil {
		t.Fatal(err)
	}
	if a.SelfMatchPct != 100 || a.QualityPct < 50 {
		t.Fatalf("unexpected analysis: %+v", a)
	}
	if !tr.Status().WinReady {
		t.Fatal("win template not ready")
	}
}

func solidRGBA(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}
