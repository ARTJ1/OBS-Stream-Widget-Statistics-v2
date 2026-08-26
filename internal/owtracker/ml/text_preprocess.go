package ml

import (
	"image"
	"image/color"

	"github.com/nfnt/resize"
)

const (
	textInputW = 128
	textInputH = 32
)

// BinaryBandFromCrop extracts gold/red banner glyphs as white-on-black and resizes for the text net.
func BinaryBandFromCrop(crop image.Image, goldPct, defeatPct int) *image.Gray {
	if crop == nil {
		return nil
	}
	band := extractTextBand(crop, goldPct, defeatPct)
	gray := binarizeBanner(band)
	if gray == nil {
		return nil
	}
	return resizeGray(textInputW, textInputH, gray)
}

func resizeGray(w, h int, src *image.Gray) *image.Gray {
	if src == nil {
		return nil
	}
	scaled := resize.Resize(uint(w), uint(h), src, resize.Lanczos3)
	out := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.GrayModel.Convert(scaled.At(x, y)).(color.Gray)
			if c.Y > 127 {
				out.SetGray(x, y, color.Gray{Y: 255})
			} else {
				out.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}
	return out
}

func grayToInput(g *image.Gray) []float32 {
	if g == nil {
		return nil
	}
	b := g.Bounds()
	w, h := b.Dx(), b.Dy()
	out := make([]float32, w*h)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if g.GrayAt(x, y).Y > 127 {
				out[i] = 1
			}
			i++
		}
	}
	return out
}

func extractTextBand(src image.Image, goldPct, defeatPct int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 8 || h < 8 {
		return src
	}
	bandH := h / 2
	if bandH < 24 {
		bandH = 24
	}
	if bandH > h {
		bandH = h
	}
	step := bandH / 4
	if step < 2 {
		step = 2
	}
	bestY, bestScore := b.Min.Y, 0
	for y := b.Min.Y; y <= b.Max.Y-bandH; y += step {
		score := scoreColorBand(src, y, y+bandH)
		if score > bestScore {
			bestScore = score
			bestY = y
		}
	}
	if bestScore < w/10 && goldPct < 12 && defeatPct < 12 {
		return src
	}
	return cropRectML(src, b.Min.X, bestY, w, bandH)
}

func scoreColorBand(img image.Image, y0, y1 int) int {
	b := img.Bounds()
	score := 0
	for y := y0; y < y1 && y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x4000 {
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			if isGoldPx(r8, g8, b8) || isRedPx(r8, g8, b8) {
				score++
			}
		}
	}
	return score
}

func binarizeBanner(src image.Image) *image.Gray {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	out := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			if a < 0x3000 {
				continue
			}
			r8, g8, b8 := int(r>>8), int(g>>8), int(bl>>8)
			if isGoldPx(r8, g8, b8) || isRedPx(r8, g8, b8) {
				out.SetGray(x, y, color.Gray{Y: 255})
			}
		}
	}
	return out
}

func isGoldPx(r, g, b int) bool {
	return r > 130 && g > 95 && b < 150 && r > b+15
}

func isRedPx(r, g, b int) bool {
	return r > 110 && r > g+18 && b < 120
}

func cropRectML(src image.Image, x, y, w, h int) image.Image {
	b := src.Bounds()
	rect := image.Rect(x, y, x+w, y+h).Intersect(b)
	if rect.Empty() {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	for dy := 0; dy < rect.Dy(); dy++ {
		for dx := 0; dx < rect.Dx(); dx++ {
			dst.Set(dx, dy, src.At(rect.Min.X+dx, rect.Min.Y+dy))
		}
	}
	return dst
}
