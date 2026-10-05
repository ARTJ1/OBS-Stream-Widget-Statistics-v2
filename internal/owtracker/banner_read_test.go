package owtracker

import (
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMatchBannerWord(t *testing.T) {
	cases := map[string]Outcome{
		"ПОБЕДА!":    OutcomeWin,
		"ПОБЕДЯ!":    OutcomeWin,
		"ПОВЕДЯ!":    OutcomeWin,
		"ПОРЯЖЕНИЕ":  OutcomeLoss,
		"VICTORY!":   OutcomeWin,
		"DEFEAT":     OutcomeLoss,
		"Défaite":    OutcomeLoss,
		"VITÓRIA!":   OutcomeWin,
		"ZWYCIĘSTWO": OutcomeWin,
		"НИЧЬЯ":      OutcomeDraw,
		"15 llfll!":  "",
		"тиш":        "",
		"ПОБЕДЫ 12":  OutcomeWin, // still a banner-ish word; color gate and size gate filter UI text
		"ELIMINATED": "",
		"OVERTIME":   "",
		// Other big in-game captions must never match.
		"РАУНД ВЫИГРАН":        "",
		"РАУНД ПРОИГРАН":       "",
		"ДОПОЛНИТЕЛЬНОЕ ВРЕМЯ": "",
		"ЛУЧШИЙ МОМЕНТ МАТЧА":  "",
		"ROUND WON":            "",
		"ROUND LOST":           "",
		"SUDDEN DEATH":         "",
		"PLAY OF THE GAME":     "",
		"ПОБЕЖДЕН":             "",
		"DEFENSE":              "",
		"":                     "",
	}
	for in, want := range cases {
		if got, _ := matchBannerWord(in); got != want {
			t.Errorf("matchBannerWord(%q) = %q, want %q", in, got, want)
		}
	}
}

// Real RU end-of-match captures (default zone at 1440p) through Windows OCR.
func TestReadBannerSamples(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows ocr only")
	}
	ocr := newWinOCR(t.TempDir())
	defer ocr.Close()
	files, _ := filepath.Glob("testdata/banners/*.png")
	if len(files) == 0 {
		t.Fatal("no samples")
	}
	for _, f := range files {
		if strings.Contains(f, "flash") {
			continue // fly-in flash: only the banner CNN is expected to read it
		}
		fh, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		img, err := decodeImage(fh)
		fh.Close()
		if err != nil {
			t.Fatal(err)
		}
		r, err := readBanner(ocr, t.TempDir(), img)
		if err != nil {
			if strings.Contains(err.Error(), "unavailable") {
				t.Skipf("windows ocr not available: %v", err)
			}
			t.Fatal(err)
		}
		name := filepath.Base(f)
		want := Outcome("")
		switch {
		case strings.HasPrefix(name, "win"):
			want = OutcomeWin
		case strings.HasPrefix(name, "loss"):
			want = OutcomeLoss
		}
		t.Logf("%s: %+v", name, r)
		if r.Outcome != want {
			t.Errorf("%s: got %q want %q (text=%q likely=%v band=%d%% fill=%d%%)",
				name, r.Outcome, want, r.Text, r.Likely, r.BandPct, r.FillPct)
		}
	}
	if langs := ocr.Langs(); len(langs) == 0 {
		t.Log("no OCR languages")
	}
}

// Full screenshots at 16:9 and 21:9: the banner zone must be found on the 16:9 area.
func TestReadBannerFullScreenshots(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows ocr only")
	}
	fh, err := os.Open("testdata/banners/loss_3.png")
	if err != nil {
		t.Fatal(err)
	}
	crop, err := decodeImage(fh)
	fh.Close()
	if err != nil {
		t.Fatal(err)
	}
	ocr := newWinOCR(t.TempDir())
	defer ocr.Close()
	for _, size := range [][2]int{{2560, 1440}, {3440, 1440}} {
		w, h := size[0], size[1]
		full := image.NewRGBA(image.Rect(0, 0, w, h))
		x0, y0, aw, ah := bannerBounds(0, 0, w, h)
		lx, ly, _, _ := defaultZone.PixelBox(aw, ah)
		draw.Draw(full, crop.Bounds().Add(image.Pt(x0+lx, y0+ly)), crop, crop.Bounds().Min, draw.Src)
		r, err := ReadBannerWith(ocr, t.TempDir(), full)
		if err != nil {
			t.Skipf("windows ocr not available: %v", err)
		}
		if r.Outcome != OutcomeLoss {
			t.Errorf("%dx%d: got %+v", w, h, r)
		}
	}
}
