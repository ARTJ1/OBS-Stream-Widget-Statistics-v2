package owtracker

import (
	"image"
	"image/color"
)

// Banner preparation for Windows OCR: the end-of-match word is huge, italic,
// condensed and glowing, which OCR cannot read as-is. We keep only gold (victory)
// or red (defeat) glyph pixels, drop decorative streaks and blobs, undo the
// italics, widen the letters and shrink to ~48 px text height (black on white).

type bannerMaskInfo struct {
	mask      []bool
	w, h      int
	bandY     int // dominant text row band
	bandH     int
	bandFill  float64 // share of glyph pixels inside the band
	gold, red int     // glyph pixel counts by color
}

// Color reported by the banner's glyph pixels.
func (m bannerMaskInfo) colorOutcome() Outcome {
	switch {
	case m.gold > 0 && m.gold >= 4*m.red:
		return OutcomeWin
	case m.red > 0 && m.red >= 4*m.gold:
		return OutcomeLoss
	}
	return ""
}

// likelyBanner is the cheap pre-check before OCR: one tall, dense, single-color text band.
func (m bannerMaskInfo) likelyBanner() bool {
	if m.h == 0 || m.bandH*100 < m.h*30 {
		return false
	}
	return m.bandFill >= 0.12 && m.colorOutcome() != ""
}

func isBannerGold(r, g, b int) bool { return r > 170 && g > 130 && b < 110 && r-b > 90 }
func isBannerRed(r, g, b int) bool  { return r > 160 && g < 70 && b < 100 && r-g > 100 }

func buildBannerMask(img image.Image) bannerMaskInfo {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	info := bannerMaskInfo{mask: make([]bool, w*h), w: w, h: h}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r16, g16, b16, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			r, g, bl := int(r16>>8), int(g16>>8), int(b16>>8)
			switch {
			case isBannerGold(r, g, bl):
				info.mask[y*w+x] = true
				info.gold++
			case isBannerRed(r, g, bl):
				info.mask[y*w+x] = true
				info.red++
			}
		}
	}
	info.mask, info.bandY, info.bandH = cleanBannerMask(info.mask, w, h)
	if info.bandH > 0 && w > 0 {
		on := 0
		for y := info.bandY; y < info.bandY+info.bandH && y < h; y++ {
			for x := 0; x < w; x++ {
				if info.mask[y*w+x] {
					on++
				}
			}
		}
		info.bandFill = float64(on) / float64(info.bandH*w)
	}
	return info
}

// cleanBannerMask keeps the dominant text row band and drops small blobs.
func cleanBannerMask(mask []bool, w, h int) ([]bool, int, int) {
	rows := make([]int, h)
	maxRow := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if mask[y*w+x] {
				rows[y]++
			}
		}
		if rows[y] > maxRow {
			maxRow = rows[y]
		}
	}
	if maxRow == 0 {
		return mask, 0, 0
	}
	bestS, bestL, s := 0, 0, -1
	for y := 0; y <= h; y++ {
		on := y < h && rows[y]*4 >= maxRow
		if on && s < 0 {
			s = y
		}
		if !on && s >= 0 {
			if y-s > bestL {
				bestS, bestL = s, y-s
			}
			s = -1
		}
	}
	pad := bestL / 12
	y0, y1 := bestS-pad, bestS+bestL+pad
	out := make([]bool, len(mask))
	for y := 0; y < h; y++ {
		if y >= y0 && y < y1 {
			copy(out[y*w:(y+1)*w], mask[y*w:(y+1)*w])
		}
	}
	minArea := bestL * bestL * 15 / 100
	seen := make([]bool, len(out))
	var stack, comp []int
	for i := range out {
		if !out[i] || seen[i] {
			continue
		}
		comp = comp[:0]
		stack = append(stack[:0], i)
		seen[i] = true
		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			comp = append(comp, p)
			x, y := p%w, p/w
			for _, q := range [4][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
				if q[0] < 0 || q[1] < 0 || q[0] >= w || q[1] >= h {
					continue
				}
				n := q[1]*w + q[0]
				if out[n] && !seen[n] {
					seen[n] = true
					stack = append(stack, n)
				}
			}
		}
		if len(comp) < minArea {
			for _, p := range comp {
				out[p] = false
			}
		}
	}
	return out, bestS, bestL
}

type bannerVariant struct {
	shear, stretch float64
}

// Tuned on real RU banners: different variants read different frames; none ever
// produced the wrong word. Tried in order until one matches.
var bannerVariants = []bannerVariant{
	{shear: 0.28, stretch: 1.5},
	{shear: 0.36, stretch: 1.4},
	{shear: 0.32, stretch: 1.3},
	{shear: 0.36, stretch: 1.3}, // streaky defeat frames at 1080p
}

const bannerTargetH = 48

// renderBannerVariant draws the cleaned mask deskewed, widened and scaled.
func renderBannerVariant(m bannerMaskInfo, v bannerVariant) *image.Gray {
	if m.w == 0 || m.h == 0 {
		return nil
	}
	scale := float64(bannerTargetH) / float64(m.h)
	sx := scale * v.stretch
	ow := int(float64(m.w)*sx+float64(bannerTargetH)*v.shear) + 40
	oh := bannerTargetH + 40
	dst := image.NewGray(image.Rect(0, 0, ow, oh))
	for i := range dst.Pix {
		dst.Pix[i] = 255
	}
	stepX := int(1/sx) + 1
	stepY := int(1/scale) + 1
	for y := 0; y < bannerTargetH; y++ {
		srcY := int(float64(y) / scale)
		for x := 0; x < ow-40; x++ {
			srcX := (float64(x) - v.shear*float64(y)) / sx
			x0 := int(srcX)
			if srcX < 0 {
				continue
			}
			cnt, on := 0, 0
			for yy := srcY; yy < srcY+stepY && yy < m.h; yy++ {
				for xx := x0; xx < x0+stepX && xx < m.w; xx++ {
					cnt++
					if m.mask[yy*m.w+xx] {
						on++
					}
				}
			}
			if cnt > 0 && on*2 >= cnt {
				dst.SetGray(x+20, y+20, color.Gray{})
			}
		}
	}
	return dst
}
