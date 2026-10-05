package owtracker

import "image"

// Stage 1 of detection: a thin full-width strip through the middle of the banner
// zone. The VICTORY/DEFEAT letters always cross it, gameplay rarely fills it with
// one banner color. Only a hit triggers the full capture + OCR (stage 2).

const (
	stripRows      = 12
	stripMinPct    = 20 // % of strip pixels in one banner color (banners: 32–72%, gameplay: 0–3%)
	stripDominance = 3  // that color must outnumber the other this many times
)

// stripColorPct counts banner gold/red pixels in BGRA rows.
func stripColorPct(bgra []byte) (goldPct, redPct int) {
	n := len(bgra) / 4
	if n == 0 {
		return 0, 0
	}
	gold, red := 0, 0
	for i := 0; i+3 < len(bgra); i += 4 {
		b, g, r := int(bgra[i]), int(bgra[i+1]), int(bgra[i+2])
		switch {
		case isBannerGold(r, g, b):
			gold++
		case isBannerRed(r, g, b):
			red++
		}
	}
	return gold * 100 / n, red * 100 / n
}

func stripLooksLikeBanner(goldPct, redPct int) bool {
	if goldPct >= stripMinPct && goldPct >= stripDominance*redPct {
		return true
	}
	return redPct >= stripMinPct && redPct >= stripDominance*goldPct
}

// stripFromImage builds the same strip from a captured image (tests, tools).
func stripFromImage(img image.Image) []byte {
	b := img.Bounds()
	y0 := b.Min.Y + b.Dy()/2 - stripRows/2
	out := make([]byte, 0, b.Dx()*stripRows*4)
	for y := y0; y < y0+stripRows; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			out = append(out, byte(bl>>8), byte(g>>8), byte(r>>8), 255)
		}
	}
	return out
}
