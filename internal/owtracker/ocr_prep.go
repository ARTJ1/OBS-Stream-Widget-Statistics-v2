package owtracker

import (
	"image"
	"image/color"
	"sort"

	"github.com/nfnt/resize"
)

type bannerMode int

const (
	bannerAuto bannerMode = iota
	bannerVictory
	bannerDefeat
)

func bannerModeFromPct(goldPct, defeatPct int) bannerMode {
	if defeatPct >= 18 && defeatPct > goldPct+10 {
		return bannerDefeat
	}
	if goldPct >= 18 && goldPct > defeatPct+10 {
		return bannerVictory
	}
	return bannerAuto
}

func cropCenterBand(src image.Image, heightFrac float64) image.Image {
	b := src.Bounds()
	h := b.Dy()
	bandH := int(float64(h) * heightFrac)
	if bandH < 24 {
		bandH = 24
	}
	if bandH > h {
		bandH = h
	}
	y := b.Min.Y + (h-bandH)/2
	return cropRect(src, b.Min.X, y, b.Dx(), bandH)
}

func prepareOCRVariants(img image.Image, goldPct, defeatPct int) []image.Image {
	if img == nil {
		return nil
	}
	mode := bannerModeFromPct(goldPct, defeatPct)
	band, ok := extractBannerTextBand(img, goldPct, defeatPct)
	if !ok {
		band = cropCenterBand(img, 0.42)
	}
	if band == nil {
		return nil
	}

	seen := map[string]struct{}{}
	var out []image.Image
	add := func(img image.Image) {
		if img == nil {
			return
		}
		up := upscaleOCR(img)
		if up == nil {
			return
		}
		key := imageKey(up)
		if _, dup := seen[key]; dup {
			return
		}
		seen[key] = struct{}{}
		out = append(out, up)
	}

	switch mode {
	case bannerDefeat:
		add(upscaleOCR(brightTextBinary(band, false)))
		add(thresholdEmphasis(redEmphasisGray(band), 0.62))
		add(erodeGray(thresholdEmphasis(redEmphasisGray(band), 0.72), 1))
		add(thresholdEmphasis(redEmphasisGray(band), 0.55))
	case bannerVictory:
		add(upscaleOCR(brightTextBinary(band, true)))
		add(thresholdEmphasis(goldEmphasisGray(band), 0.62))
		add(erodeGray(thresholdEmphasis(goldEmphasisGray(band), 0.72), 1))
		add(thresholdEmphasis(goldEmphasisGray(band), 0.55))
	default:
		add(upscaleOCR(brightTextBinary(band, true)))
		add(upscaleOCR(brightTextBinary(band, false)))
		add(thresholdEmphasis(redEmphasisGray(band), 0.65))
		add(thresholdEmphasis(goldEmphasisGray(band), 0.65))
	}
	add(toHighContrastGray(band))
	add(upscaleOCR(band))

	if len(out) == 0 {
		add(upscaleOCR(toHighContrastGray(band)))
	}
	return out
}

func prepareOCRImage(img image.Image, goldPct, defeatPct int) (image.Image, bool) {
	variants := prepareOCRVariants(img, goldPct, defeatPct)
	if len(variants) == 0 {
		return nil, false
	}
	return variants[0], true
}

func upscaleOCR(img image.Image) image.Image {
	if img == nil {
		return nil
	}
	b := img.Bounds()
	w := b.Dx()
	minW := 1400
	if w >= minW {
		return img
	}
	scale := float64(minW) / float64(w)
	nw := int(float64(w) * scale)
	nh := int(float64(b.Dy()) * scale)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	return resize.Resize(uint(nw), uint(nh), img, resize.Lanczos3)
}

func imageKey(img image.Image) string {
	b := img.Bounds()
	if b.Empty() {
		return ""
	}
	r, g, bl, _ := img.At(b.Min.X, b.Min.Y).RGBA()
	r2, g2, b2, _ := img.At(b.Max.X-1, b.Max.Y-1).RGBA()
	return string([]byte{
		byte(r >> 8), byte(g >> 8), byte(bl >> 8),
		byte(r2 >> 8), byte(g2 >> 8), byte(b2 >> 8),
		byte(b.Dx()), byte(b.Dy()),
	})
}

func redEmphasisGray(src image.Image) *image.Gray {
	b := src.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			if a < 0x3000 {
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			m := g8
			if b8 > m {
				m = b8
			}
			v := r8 - m - 20
			if v < 0 {
				v = 0
			}
			if isBannerRedPixel(r8, g8, b8) {
				v += 40
			}
			if v > 255 {
				v = 255
			}
			out.SetGray(x, y, color.Gray{Y: uint8(v)})
		}
	}
	return out
}

func goldEmphasisGray(src image.Image) *image.Gray {
	b := src.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			if a < 0x3000 {
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			v := (r8 + g8) / 2
			if b8 > v-20 {
				v -= b8 / 3
			}
			if isBannerGoldPixel(r8, g8, b8) {
				v += 50
			}
			if v < 0 {
				v = 0
			}
			if v > 255 {
				v = 255
			}
			out.SetGray(x, y, color.Gray{Y: uint8(v)})
		}
	}
	return out
}

func thresholdEmphasis(src *image.Gray, pct float64) *image.Gray {
	if src == nil {
		return image.NewGray(image.Rect(0, 0, 1, 1))
	}
	b := src.Bounds()
	vals := make([]int, 0, b.Dx()*b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			v := int(src.GrayAt(x, y).Y)
			if v > 8 {
				vals = append(vals, v)
			}
		}
	}
	out := image.NewGray(b)
	if len(vals) == 0 {
		return out
	}
	sort.Ints(vals)
	th := vals[int(float64(len(vals)-1)*pct)]
	if th < 40 {
		th = 40
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if int(src.GrayAt(x, y).Y) >= th {
				out.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}
	return out
}

func erodeGray(src *image.Gray, passes int) *image.Gray {
	if src == nil || passes <= 0 {
		return src
	}
	cur := src
	for p := 0; p < passes; p++ {
		b := cur.Bounds()
		next := image.NewGray(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				if cur.GrayAt(x, y).Y < 200 {
					continue
				}
				keep := true
				for dy := -1; dy <= 1 && keep; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if cur.GrayAt(x+dx, y+dy).Y < 200 {
							keep = false
							break
						}
					}
				}
				if keep {
					next.SetGray(x, y, color.Gray{Y: 255})
				}
			}
		}
		cur = next
	}
	return cur
}

// brightTextBinary isolates large gold/red banner letters on a dark background for OCR.
func brightTextBinary(src image.Image, goldPreferred bool) *image.Gray {
	b := src.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			if a < 0x3000 {
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			lum := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)
			on := false
			if goldPreferred {
				on = isBannerGoldPixel(r8, g8, b8) || lum >= 175
			} else {
				on = isBannerRedPixel(r8, g8, b8) || lum >= 175
			}
			if on {
				out.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}
	return out
}

// strongColorBannerOutcome is a last resort when OCR fails but banner colors are unambiguous.
func strongColorBannerOutcome(goldPct, defeatPct int) string {
	return strongColorBannerOutcomeForLuma(goldPct, defeatPct, -1)
}

func strongColorBannerOutcomeForLuma(goldPct, defeatPct, avgLum int) string {
	if defeatPct >= 26 && defeatPct > goldPct+16 {
		return "loss"
	}
	if goldPct >= 26 && goldPct > defeatPct+16 {
		return "win"
	}
	// Dark end-screen background: text is small % of crop but clearly red/gold.
	if avgLum >= 0 && avgLum <= 55 {
		if defeatPct >= 16 && defeatPct > goldPct+8 {
			return "loss"
		}
		if goldPct >= 16 && goldPct > defeatPct+8 {
			return "win"
		}
	}
	return ""
}
