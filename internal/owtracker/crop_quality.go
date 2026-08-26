package owtracker

import "image"

// cropAvgLuminance returns mean brightness 0–255 for the crop.
func cropAvgLuminance(img image.Image) int {
	return avgLuminance(img)
}

// isBlankCrop detects failed captures (all black/empty) or fully white frames.
func isBlankCrop(img image.Image) bool {
	if img == nil {
		return true
	}
	b := img.Bounds()
	if b.Empty() {
		return true
	}
	avg := cropAvgLuminance(img)
	if avg <= 6 || avg >= 252 {
		return true
	}
	dark := 0
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x3000 {
				continue
			}
			n++
			v := int(0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8))
			if v <= 12 {
				dark++
			}
		}
	}
	if n == 0 {
		return true
	}
	return dark*100/n >= 97
}

func max8(a, b int) int {
	if a > b {
		return a
	}
	return b
}
