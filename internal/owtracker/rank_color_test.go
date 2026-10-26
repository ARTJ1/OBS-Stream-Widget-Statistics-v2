package owtracker

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/nfnt/resize"
)

// Color changes the reader must survive, on the real progress screens of both
// clients: the emblem glow is a different color for every tier (hue turns), the
// streamer's OBS color correction or a game filter (tints, gamma, saturation),
// Overwatch's color blind filters, a dimmed screen and a low quality stream.
var rankColorVariants = []struct {
	name string
	fn   func(r, g, b float64) (float64, float64, float64)
}{
	{"hue60", hueTurn(60)}, {"hue120", hueTurn(120)}, {"hue180", hueTurn(180)},
	{"hue240", hueTurn(240)}, {"hue300", hueTurn(300)},
	{"warm", scaleRGB(1.0, 0.88, 0.7)}, {"cool", scaleRGB(0.75, 0.9, 1.0)},
	{"green", scaleRGB(0.8, 1.0, 0.8)}, {"purple", scaleRGB(0.95, 0.75, 1.0)},
	{"dark", gammaRGB(1.6, 0.75)}, {"bright", gammaRGB(0.7, 1.0)},
	{"vivid", saturate(1.6)}, {"gray", saturate(0)},
	{"protanopia", mix3(0.152, 1.053, -0.205, 0.115, 0.786, 0.099, -0.004, -0.048, 1.052)},
	{"deuteranopia", mix3(0.367, 0.861, -0.228, 0.280, 0.673, 0.047, -0.012, 0.043, 0.969)},
	{"tritanopia", mix3(1.256, -0.077, -0.179, -0.078, 0.931, 0.148, 0.005, 0.691, 0.304)},
}

func TestReadRankScreenColors(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows ocr only")
	}
	if testing.Short() {
		t.Skip("slow: ~200 OCR reads")
	}
	ocr := newWinOCR(t.TempDir())
	defer ocr.Close()
	langs := rankTestLangs(t, ocr)
	files, _ := filepath.Glob("testdata/rank/*.jpg")
	bad, total := 0, 0
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".jpg")
		parts := strings.Split(name, "_")
		if parts[0] == "none" || !langs[parts[0]] {
			continue
		}
		fh, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		src, err := decodeImage(fh)
		fh.Close()
		if err != nil {
			t.Fatal(err)
		}
		div, _ := strconv.Atoi(parts[3])
		want := RankRead{Role: parts[1], Tier: parts[2], Division: div, Rank: RankIndex(parts[2], div)}
		variants := append(rankColorVariants, struct {
			name string
			fn   func(r, g, b float64) (float64, float64, float64)
		}{"jpeg40", nil})
		for _, v := range variants {
			var img image.Image
			if v.fn == nil {
				img = jpegRoundTrip(t, src, 40)
			} else {
				img = recolor(src, v.fn)
			}
			total++
			if !rankGate(cropRankGate(resize.Resize(480, 270, img, resize.Bilinear))) {
				t.Errorf("%s/%s: gate missed the screen", name, v.name)
			}
			r, err := ReadRankScreen(ocr, t.TempDir(), img)
			if err != nil {
				t.Fatal(err)
			}
			got := RankRead{Role: r.Role, Tier: r.Tier, Division: r.Division, Rank: r.Rank}
			if got != want {
				bad++
				t.Errorf("%s/%s: got %+v (role %q, label %q)", name, v.name, got, r.RoleText, r.RankText)
			}
		}
	}
	t.Logf("%d/%d color variants wrong", bad, total)
}

func recolor(src image.Image, fn func(r, g, b float64) (float64, float64, float64)) image.Image {
	b := src.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := src.At(x, y).RGBA()
			nr, ng, nb := fn(float64(r>>8)/255, float64(g>>8)/255, float64(bb>>8)/255)
			out.SetRGBA(x, y, color.RGBA{clamp8(nr), clamp8(ng), clamp8(nb), 255})
		}
	}
	return out
}

func clamp8(v float64) uint8 { return uint8(math.Round(math.Max(0, math.Min(1, v)) * 255)) }

func scaleRGB(kr, kg, kb float64) func(r, g, b float64) (float64, float64, float64) {
	return func(r, g, b float64) (float64, float64, float64) { return r * kr, g * kg, b * kb }
}

func gammaRGB(gamma, gain float64) func(r, g, b float64) (float64, float64, float64) {
	f := func(v float64) float64 { return gain * math.Pow(v, gamma) }
	return func(r, g, b float64) (float64, float64, float64) { return f(r), f(g), f(b) }
}

func saturate(s float64) func(r, g, b float64) (float64, float64, float64) {
	return func(r, g, b float64) (float64, float64, float64) {
		l := 0.299*r + 0.587*g + 0.114*b
		return l + (r-l)*s, l + (g-l)*s, l + (b-l)*s
	}
}

func mix3(m ...float64) func(r, g, b float64) (float64, float64, float64) {
	return func(r, g, b float64) (float64, float64, float64) {
		return m[0]*r + m[1]*g + m[2]*b, m[3]*r + m[4]*g + m[5]*b, m[6]*r + m[7]*g + m[8]*b
	}
}

// hueTurn rotates hue and keeps grays gray (the CSS hue-rotate matrix).
func hueTurn(deg float64) func(r, g, b float64) (float64, float64, float64) {
	c, s := math.Cos(deg*math.Pi/180), math.Sin(deg*math.Pi/180)
	return mix3(
		0.213+c*0.787-s*0.213, 0.715-c*0.715-s*0.715, 0.072-c*0.072+s*0.928,
		0.213-c*0.213+s*0.143, 0.715+c*0.285+s*0.140, 0.072-c*0.072-s*0.283,
		0.213-c*0.213-s*0.787, 0.715-c*0.715+s*0.715, 0.072+c*0.928+s*0.072,
	)
}

func jpegRoundTrip(t *testing.T, src image.Image, q int) image.Image {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: q}); err != nil {
		t.Fatal(err)
	}
	img, err := jpeg.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	return img
}
