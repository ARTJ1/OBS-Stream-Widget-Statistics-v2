package owtracker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBannerShapeWinTemplate(t *testing.T) {
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
	gold, defeat := bannerColorHint(img)
	r := ProbeResult{
		EndScreenActive: true,
		HashEndScreen:   true,
		GoldPct:         gold,
		DefeatPct:       defeat,
	}
	match, ok := bannerShapeOutcome(img, r)
	if !ok || match != "win" {
		t.Fatalf("expected shape win, got %q ok=%v gold=%d defeat=%d", match, ok, gold, defeat)
	}
}

func TestBannerShapeRejectsHashOnly(t *testing.T) {
	winPath := filepath.Join("..", "..", "internal", "owtracker", "assets", "templates", "win.png")
	f, _ := os.Open(winPath)
	img, _ := decodeImage(f)
	f.Close()
	r := ProbeResult{
		EndScreenActive: true,
		HashEndScreen:   true,
		GoldPct:         2,
		DefeatPct:       4,
		WinSimilarity:   63,
		LossSimilarity:  50,
		WinDistance:     20,
		LossDistance:    28,
	}
	_, ok := endScreenBannerFallback(r, img, 20, 28, 6, 8, 8)
	if ok {
		t.Fatal("hash-like weak colors must not trigger")
	}
}

func TestApplyProbeAllowedRequiresColors(t *testing.T) {
	if applyProbeAllowed(ProbeResult{Match: "win", MatchMethod: "banner", EndScreenActive: true, GoldPct: 2, DefeatPct: 4}) {
		t.Fatal("weak gold should reject")
	}
	if !applyProbeAllowed(ProbeResult{Match: "win", MatchMethod: "ml", EndScreenActive: true, GoldPct: 30, DefeatPct: 8}) {
		t.Fatal("strong gold should allow")
	}
}
