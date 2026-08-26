package owtracker

import (
	"image"
	"sync"
	"time"
)

const endScreenWindow = 45 * time.Second

// endScreenWatch opens a short window for OCR only after a real end-screen banner is seen.
type endScreenWatch struct {
	mu         sync.Mutex
	wasInMatch bool
	hadHUD     bool
	active     bool
	until      time.Time
	pending    int
	miss       int
}

func (w *endScreenWatch) tick(hudDetected bool, crop image.Image, testMode bool, hashEndScreen bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	gold, defeat := bannerColorHint(crop)
	signal := hashEndScreen || endScreenBanner(gold, defeat) || endScreenBannerImage(crop)

	if hudDetected {
		w.wasInMatch = true
		w.hadHUD = true
	}

	// After HUD disappears, only a real banner shape opens OCR — not red MVP nameplates.
	if !signal && w.wasInMatch && w.hadHUD && !hudDetected {
		signal = endScreenBannerImage(crop)
	}

	// Production: strong banner alone is enough (HUD can be missed after UI updates).
	if !testMode && !w.wasInMatch {
		signal = endScreenBanner(gold, defeat) || endScreenBannerImage(crop)
	}

	if signal {
		w.pending++
		w.miss = 0
	} else {
		if w.active {
			w.miss++
		} else {
			w.pending = 0
		}
	}

	activate := w.pending >= 2 || strongBannerColor(gold, defeat) || hashEndScreen
	if activate && (signal || hashEndScreen) {
		w.active = true
		w.until = time.Now().Add(endScreenWindow)
	}

	if w.active && w.miss >= 8 {
		w.active = false
		w.pending = 0
		w.miss = 0
	}

	if w.active && time.Now().After(w.until) {
		w.active = false
		w.wasInMatch = false
		w.hadHUD = false
		w.pending = 0
		w.miss = 0
	}

	// Keep window open while a strong banner is visible (stream flicker).
	if strongBannerColor(gold, defeat) {
		w.active = true
		w.until = time.Now().Add(endScreenWindow)
		w.miss = 0
	}
}

// endScreenBanner detects large ПОБЕДА/ПОРАЖЕНИЕ color blocks, not generic dark/red gameplay UI.
func endScreenBanner(gold, defeat int) bool {
	if gold >= 25 && gold > defeat+15 {
		return true
	}
	if defeat >= 25 && defeat > gold+15 {
		return true
	}
	return false
}

// endScreenBannerImage requires a wide horizontal gold/red glyph band (banner word), not scattered UI.
func endScreenBannerImage(img image.Image) bool {
	if img == nil {
		return false
	}
	gold, defeat := bannerColorHint(img)
	band, ok := extractBannerTextBand(img, gold, defeat)
	if !ok {
		if defeat >= 26 && defeat > gold+16 {
			return true
		}
		if gold >= 26 && gold > defeat+16 {
			return true
		}
		return false
	}
	return bannerBinaryScore(bannerBinary(band)) >= 80
}

func strongBannerColor(gold, defeat int) bool {
	if defeat >= 28 && defeat > gold+16 {
		return true
	}
	if gold >= 28 && gold > defeat+16 {
		return true
	}
	return false
}

func (w *endScreenWatch) isPending() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.pending > 0 || w.active
}

func (w *endScreenWatch) isActive() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.active && time.Now().After(w.until) {
		w.active = false
	}
	return w.active
}

func (w *endScreenWatch) reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.wasInMatch = false
	w.hadHUD = false
	w.active = false
	w.pending = 0
	w.miss = 0
}

func avgLuminance(img image.Image) int {
	if img == nil {
		return 255
	}
	b := img.Bounds()
	n := 0
	var sum float64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a < 0x4000 {
				continue
			}
			n++
			sum += 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8)
		}
	}
	if n == 0 {
		return 255
	}
	return int(sum / float64(n))
}
