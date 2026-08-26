package owtracker

import (
	"image"
	"strings"
)

const (
	CaptureScreen = "screen"
	CaptureWindow = "window"
	CaptureBoth   = "both"
)

const matchDiscrimination = 4

func centerBox(boundW, boundH int) (x, y, w, h int) {
	w = boundW * 25 / 100
	h = boundH * 10 / 100
	if w < 8 {
		w = 8
	}
	if h < 8 {
		h = 8
	}
	x = boundW/2 - w/2
	y = boundH/2 - h/2
	return x, y, w, h
}

func hudBox(boundW, boundH int) (x, y, w, h int) {
	w, h = 80, 48
	x = boundW/2 - w/2
	y = boundH * 2 / 100
	if y < 0 {
		y = 0
	}
	return x, y, w, h
}

func captureGameZone(zone Zone) (image.Image, error) {
	x0, y0, bw, bh, ok := overwatchClientBounds()
	if !ok {
		return nil, ErrGameNotVisible
	}
	lx, ly, w, h := zone.PixelBox(bw, bh)
	return captureRegion(x0+lx, y0+ly, w, h)
}

func captureScreenZone(zone Zone) (image.Image, string, error) {
	sw, sh := screenSize()
	if sw <= 0 || sh <= 0 {
		return nil, "", ErrGameNotVisible
	}
	lx, ly, w, h := zone.PixelBox(sw, sh)
	img, err := captureRegion(lx, ly, w, h)
	return img, "(screen)", err
}

func captureWindowZone(zone Zone) (image.Image, string, error) {
	x0, y0, bw, bh, title, ok := foregroundClientBoundsWithTitle()
	if !ok {
		return captureScreenZone(zone)
	}
	lx, ly, w, h := zone.PixelBox(bw, bh)
	img, err := captureRegion(x0+lx, y0+ly, w, h)
	return img, title, err
}

func captureHUD() (image.Image, error) {
	x0, y0, bw, bh, ok := overwatchClientBounds()
	if !ok {
		return nil, ErrGameNotVisible
	}
	lx, ly, w, h := hudBox(bw, bh)
	return captureRegion(x0+lx, y0+ly, w, h)
}

func captureHUDTest() (image.Image, error) {
	x0, y0, bw, bh, ok := foregroundClientBounds()
	if !ok {
		return nil, ErrGameNotVisible
	}
	lx, ly, w, h := hudBox(bw, bh)
	return captureRegion(x0+lx, y0+ly, w, h)
}

func isOverwatchTitle(title string) bool {
	t := strings.TrimSpace(title)
	if t == "" {
		return false
	}
	if t == "Overwatch" || t == "Overwatch 2" {
		return true
	}
	return strings.HasPrefix(strings.ToLower(t), "overwatch")
}
