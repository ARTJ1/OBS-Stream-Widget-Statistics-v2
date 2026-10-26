package owtracker

import "image"

// Cheap look for the progress screen, so the auto rank reader also runs when the
// banner was missed (a draw, a banner hidden behind a menu, a quick leave)
// without OCR on every frame. The role header ("+ SUPPORT") is short gray-white
// text centered on the screen with nothing bright beside it, also when the screen
// is dimmed under the endorse cards. The check looks at one thin band across the
// middle: from the 480 px banner frame (OBS) or a small screen capture.

// rankGateBand: the role header row, x0, y0, x1, y1 (fractions of the 16:9 area).
var rankGateBand = [4]float64{0.33, 0.175, 0.67, 0.225}

const (
	rankGateMinText = 0.03 // the word covers some of the middle …
	rankGateMaxText = 0.70 // … but is not a bright patch
	rankGateMaxSide = 0.04 // and there is next to nothing beside it
)

// rankGateRect is the band inside a w x h game picture.
func rankGateRect(w, h int) image.Rectangle {
	x, y, aw, ah := bannerBounds(0, 0, w, h)
	return image.Rect(
		x+int(rankGateBand[0]*float64(aw)), y+int(rankGateBand[1]*float64(ah)),
		x+int(rankGateBand[2]*float64(aw)), y+int(rankGateBand[3]*float64(ah)),
	)
}

// cropRankGate cuts the band out of a full game frame.
func cropRankGate(img image.Image) image.Image {
	b := img.Bounds()
	r := rankGateRect(b.Dx(), b.Dy()).Add(b.Min)
	if sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		return sub.SubImage(r)
	}
	return nil
}

// rankGateStats: share of text pixels in the middle of the band and beside it.
// Text is gray-white and clearly brighter than the band around it.
func rankGateStats(band image.Image) (mid, side float64) {
	b := band.Bounds()
	if b.Dx() < 40 || b.Dy() < 4 {
		return 0, 1
	}
	// The band spans 0.33–0.67 of the screen: the text lives in 0.43–0.57, the
	// sides 0.33–0.41 and 0.59–0.67 must stay dark.
	span := rankGateBand[2] - rankGateBand[0]
	col := func(f float64) int { return b.Min.X + int((f-rankGateBand[0])/span*float64(b.Dx())) }
	cols := [3][2]int{{col(0.33), col(0.41)}, {col(0.43), col(0.57)}, {col(0.59), col(0.67)}}
	lum := func(px, py int) (int, bool) {
		r, g, bb, _ := band.At(px, py).RGBA()
		r8, g8, b8 := int(r>>8), int(g>>8), int(bb>>8)
		// Near-gray relative to brightness, so tinted white text still counts.
		mx := max(r8, g8, b8)
		return (r8*3 + g8*6 + b8) / 10, (mx-min(r8, g8, b8))*100 <= 30*mx+10
	}
	var hist [256]int
	n := 0
	for _, c := range cols {
		for py := b.Min.Y; py < b.Max.Y; py++ {
			for px := c[0]; px < c[1]; px++ {
				l, _ := lum(px, py)
				hist[l]++
				n++
			}
		}
	}
	med, acc := 0, 0
	for med < 255 && acc+hist[med] < n/2 {
		acc += hist[med]
		med++
	}
	thr := max(60, med+45)
	var nMid, nSide, aMid, aSide int
	for i, c := range cols {
		for py := b.Min.Y; py < b.Max.Y; py++ {
			for px := c[0]; px < c[1]; px++ {
				l, gray := lum(px, py)
				text := gray && l >= thr
				if i == 1 {
					aMid++
					if text {
						nMid++
					}
				} else {
					aSide++
					if text {
						nSide++
					}
				}
			}
		}
	}
	if aMid == 0 || aSide == 0 {
		return 0, 1
	}
	return float64(nMid) / float64(aMid), float64(nSide) / float64(aSide)
}

// rankGate reports whether the band may be the progress screen's role header.
func rankGate(band image.Image) bool {
	if band == nil {
		return false
	}
	mid, side := rankGateStats(band)
	return mid >= rankGateMinText && mid <= rankGateMaxText && side <= rankGateMaxSide
}
