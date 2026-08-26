//go:build windows

package owtracker

import (
	"image"
	"strings"
	"syscall"
)

func isStreamBrowserTitle(title string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	if t == "" {
		return false
	}
	hasStream := strings.Contains(t, "twitch") || strings.Contains(t, "youtube")
	if !hasStream {
		return false
	}
	return strings.Contains(t, "opera") ||
		strings.Contains(t, "chrome") ||
		strings.Contains(t, "firefox") ||
		strings.Contains(t, "edge") ||
		strings.Contains(t, "brave")
}

func isIgnoredTestCaptureTitle(title string) bool {
	t := strings.ToLower(strings.TrimSpace(title))
	if t == "" {
		return false
	}
	if strings.Contains(t, "widget stats") {
		return true
	}
	if strings.Contains(t, "admin") && strings.Contains(t, "chrome") {
		return true
	}
	return false
}

func findStreamBrowserHWND() uintptr {
	type cand struct {
		hwnd uintptr
		area int
	}
	var best cand
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		title := windowTitle(hwnd)
		if !isStreamBrowserTitle(title) {
			return 1
		}
		_, _, w, h, ok := clientBoundsOf(hwnd)
		if !ok {
			return 1
		}
		area := w * h
		if area > best.area {
			best = cand{hwnd: hwnd, area: area}
		}
		return 1
	})
	_, _, _ = procEnumWindows.Call(cb, 0)
	return best.hwnd
}

func captureClientZoneOf(hwnd uintptr, zone Zone) (image.Image, string, error) {
	x0, y0, bw, bh, ok := clientBoundsOf(hwnd)
	if !ok {
		return nil, "", ErrGameNotVisible
	}
	lx, ly, w, h := zone.PixelBox(bw, bh)
	img, err := captureRegion(x0+lx, y0+ly, w, h)
	return img, windowTitle(hwnd), err
}

// captureWindowZoneSmart prefers Opera/Twitch even when Admin panel is foreground.
func captureWindowZoneSmart(zone Zone) (image.Image, string, error) {
	if x0, y0, bw, bh, title, ok := foregroundClientBoundsWithTitle(); ok {
		if !isIgnoredTestCaptureTitle(title) {
			lx, ly, w, h := zone.PixelBox(bw, bh)
			img, err := captureRegion(x0+lx, y0+ly, w, h)
			return img, title, err
		}
	}
	if hwnd := findStreamBrowserHWND(); hwnd != 0 {
		return captureClientZoneOf(hwnd, zone)
	}
	return captureWindowZone(zone)
}
