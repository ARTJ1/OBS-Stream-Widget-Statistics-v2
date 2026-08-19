package owtracker

import (
	"image"
)

// looksLikeMatchHUD reports whether a small top-center crop contains Overwatch 2
// in-match chrome (timer digits and/or blue/orange round pips).
func looksLikeMatchHUD(img image.Image) bool {
	if img == nil {
		return false
	}
	b := img.Bounds()
	total := b.Dx() * b.Dy()
	if total <= 0 {
		return false
	}
	var ui, white int
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x8000 {
				continue
			}
			rr := uint8(r >> 8)
			gg := uint8(g >> 8)
			bb := uint8(bl >> 8)
			if isHUDWhite(rr, gg, bb) {
				white++
				ui++
				continue
			}
			if isTeamBlue(rr, gg, bb) || isEnemyOrange(rr, gg, bb) {
				ui++
			}
		}
	}
	// 40×40 crop: a handful of timer/pip pixels is enough; menus are mostly dark/desaturated.
	return ui >= total/12 && white >= 4
}

func isHUDWhite(r, g, b uint8) bool {
	if r < 200 || g < 200 || b < 180 {
		return false
	}
	mx, mn := max3(r, g, b), min3(r, g, b)
	return int(mx)-int(mn) < 50
}

func isTeamBlue(r, g, b uint8) bool {
	return b > 140 && b > r+30 && b > g && g > 80
}

func isEnemyOrange(r, g, b uint8) bool {
	return r > 160 && g > 50 && g < 180 && b < 90 && r > g && r > b+40
}

func max3(a, b, c uint8) uint8 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}

func min3(a, b, c uint8) uint8 {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
