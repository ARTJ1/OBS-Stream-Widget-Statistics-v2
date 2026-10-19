package owtracker

import (
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/nfnt/resize"
)

// Real 1280x720 frames from recorded streams (the size the widget asks OBS for)
// through Windows OCR. Names give the expected read:
//
//	<lang>_<role>_<Tier>_<division>[_note].jpg  progress screen, must read exactly
//	none_<lang>_<what>.jpg                      anything else, must read nothing
//
// The none_ frames are traps seen next to the progress screen: other players'
// endorsement cards with their ranks, the scoreboard ("DAMAGE MITIGATED"), the
// menu role cards and the career "Competitive progress" page.
func TestReadRankScreenSamples(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows ocr only")
	}
	ocr := newWinOCR(t.TempDir())
	defer ocr.Close()
	langs := rankTestLangs(t, ocr)
	files, _ := filepath.Glob("testdata/rank/*.jpg")
	if len(files) < 10 {
		t.Fatalf("only %d samples", len(files))
	}
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".jpg")
		parts := strings.Split(name, "_")
		lang := parts[0]
		if lang == "none" {
			lang = parts[1]
		}
		if !langs[lang] {
			t.Logf("%s: skipped, no %q Windows OCR language", name, lang)
			continue
		}
		t.Run(name, func(t *testing.T) {
			fh, err := os.Open(f)
			if err != nil {
				t.Fatal(err)
			}
			img, err := decodeImage(fh)
			fh.Close()
			if err != nil {
				t.Fatal(err)
			}
			r, err := ReadRankScreen(ocr, t.TempDir(), img)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("read %+v", r)
			if parts[0] == "none" {
				if r.Valid() {
					t.Errorf("trap frame read as a rank: %+v", r)
				}
				return
			}
			div, _ := strconv.Atoi(parts[3])
			want := RankRead{Role: parts[1], Tier: parts[2], Division: div, Rank: RankIndex(parts[2], div)}
			got := RankRead{Role: r.Role, Tier: r.Tier, Division: r.Division, Rank: r.Rank}
			if got != want {
				t.Errorf("got %+v (role %q, label %q), want %+v", got, r.RoleText, r.RankText, want)
			}
		})
	}
}

// Other client sizes (ultrawide, 16:10, native 1080p/1440p screen capture): the
// reader must find the 16:9 area like the banner reader.
func TestReadRankScreenAspectRatios(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows ocr only")
	}
	ocr := newWinOCR(t.TempDir())
	defer ocr.Close()
	if !rankTestLangs(t, ocr)["ru"] {
		t.Skip("no Russian Windows OCR language")
	}
	fh, err := os.Open("testdata/rank/ru_damage_Emerald_2.jpg")
	if err != nil {
		t.Fatal(err)
	}
	src, err := decodeImage(fh)
	fh.Close()
	if err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int{{1720, 720}, {1280, 800}, {1920, 1080}, {2560, 1440}, {3440, 1440}} {
		w, h := size[0], size[1]
		// Native-size screen capture: scale the 16:9 frame to the client height.
		fw, fh := h*16/9, h
		if fw > w {
			fw, fh = w, w*9/16
		}
		frame := resize.Resize(uint(fw), uint(fh), src, resize.Lanczos3)
		r, err := ReadRankScreen(ocr, t.TempDir(), padTo16x9(frame, w, h))
		if err != nil {
			t.Fatal(err)
		}
		if r.Role != "damage" || r.Tier != "Emerald" || r.Division != 2 {
			t.Errorf("%dx%d: got %+v", size[0], size[1], r)
		}
	}
}

func rankTestLangs(t *testing.T, o *winOCR) map[string]bool {
	t.Helper()
	o.mu.Lock()
	err := o.start()
	langs := append([]string(nil), o.langs...)
	o.mu.Unlock()
	if err != nil {
		t.Skipf("windows ocr not available: %v", err)
	}
	m := map[string]bool{}
	for _, l := range langs {
		m[l] = true
	}
	if !m["en"] && !m["ru"] {
		t.Skipf("no en/ru Windows OCR language (have %v)", langs)
	}
	return m
}

// padTo16x9 centers a 16:9 frame on a black w x h canvas (letterbox/pillarbox).
func padTo16x9(src image.Image, w, h int) image.Image {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(out, out.Bounds(), image.Black, image.Point{}, draw.Src)
	b := src.Bounds()
	at := image.Pt((w-b.Dx())/2, (h-b.Dy())/2)
	draw.Draw(out, image.Rectangle{Min: at, Max: at.Add(b.Size())}, src, b.Min, draw.Src)
	return out
}
