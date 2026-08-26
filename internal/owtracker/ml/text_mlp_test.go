package ml_test

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"
)

func TestTextNetClassifiesWinBinary(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	ml.InitText(filepath.Join(root, "data"))
	if !ml.TextReady() {
		t.Skip("text model not loaded")
	}
	f, err := os.Open(filepath.Join(root, "internal", "owtracker", "assets", "templates", "win.png"))
	if err != nil {
		t.Skip(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ml.ClassifyBannerText(img, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Label != ml.LabelWin {
		t.Fatalf("expected win, got %s win=%.2f loss=%.2f none=%.2f", r.Label, r.WinProb, r.LossProb, r.NoneProb)
	}
}
