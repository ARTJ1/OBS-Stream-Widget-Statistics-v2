package owtracker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEndScreenFallbackWinOnGoldBanner(t *testing.T) {
	path := filepath.Join("..", "..", "internal", "owtracker", "assets", "templates", "win.png")
	f, err := os.Open(path)
	if err != nil {
		t.Skip(err)
	}
	img, err := decodeImage(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	tr := New(t.TempDir(), nil)
	tr.testMode.Store(true)
	tr.hashes.setWin("p:c01f6b0eb8477c63")
	tr.hashes.setLoss("p:91a07af02f3d55c3")
	r := tr.probeCrop(img, img.Bounds().Dx(), img.Bounds().Dy(), "test", false, "preview")
	if !r.WouldTrigger || r.Match != "win" {
		t.Fatalf("expected win trigger, got match=%q trigger=%v notes=%q gold=%d defeat=%d",
			r.Match, r.WouldTrigger, r.Notes, r.GoldPct, r.DefeatPct)
	}
}
