package owtracker

import (
	"image"
	"math"
)

// Zone is a capture region as percentages of the game window or screen (0–100).
type Zone struct {
	XPct float64 `json:"xPct"`
	YPct float64 `json:"yPct"`
	WPct float64 `json:"wPct"`
	HPct float64 `json:"hPct"`
}

func (z Zone) Valid() bool {
	if z.WPct < 1 || z.HPct < 1 {
		return false
	}
	if z.XPct < 0 || z.YPct < 0 {
		return false
	}
	return z.XPct+z.WPct <= 100.01 && z.YPct+z.HPct <= 100.01
}

func (z Zone) PixelBox(boundW, boundH int) (x, y, w, h int) {
	if boundW <= 0 || boundH <= 0 {
		return 0, 0, 0, 0
	}
	w = int(math.Round(float64(boundW) * z.WPct / 100))
	h = int(math.Round(float64(boundH) * z.HPct / 100))
	x = int(math.Round(float64(boundW) * z.XPct / 100))
	y = int(math.Round(float64(boundH) * z.YPct / 100))
	if w < 8 {
		w = 8
	}
	if h < 8 {
		h = 8
	}
	if x+w > boundW {
		x = boundW - w
	}
	if y+h > boundH {
		y = boundH - h
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y, w, h
}

func cropZone(src image.Image, zone Zone) (image.Image, bool) {
	if !zone.Valid() {
		return nil, false
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	x, y, cw, ch := zone.PixelBox(w, h)
	if cw <= 0 || ch <= 0 {
		return nil, false
	}
	return cropRGBA(src, b.Min.X+x, b.Min.Y+y, cw, ch), true
}

func parseZone(xPct, yPct, wPct, hPct float64) Zone {
	return Zone{XPct: xPct, YPct: yPct, WPct: wPct, HPct: hPct}
}

// DefaultZone is only a starting rectangle in the ROI editor (user must confirm).
func DefaultZone() Zone {
	return Zone{XPct: 37.5, YPct: 45, WPct: 25, HPct: 10}
}
