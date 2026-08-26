package ml_test

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"
)

func loadPNG(t *testing.T, path string) image.Image {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func TestMLPClassifiesWinTemplate(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	ml.Init(filepath.Join(root, "data"))
	if !ml.Ready() {
		t.Skip("ml model not loaded")
	}
	img := loadPNG(t, filepath.Join(root, "internal", "owtracker", "assets", "templates", "win.png"))
	r, err := ml.Classify(img)
	if err != nil {
		t.Fatal(err)
	}
	if r.Label != ml.LabelWin {
		t.Fatalf("expected win, got %s conf=%.2f", r.Label, r.Confidence)
	}
}
