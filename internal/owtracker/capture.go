package owtracker

import (
	"image"
	"strings"
)

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
	w, h = 40, 40
	x = boundW/2 - w/2
	y = boundH * 2 / 100
	if y < 0 {
		y = 0
	}
	return x, y, w, h
}

func captureCenter() (image.Image, error) {
	x0, y0, bw, bh, ok := overwatchClientBounds()
	if !ok {
		return nil, ErrGameNotVisible
	}
	lx, ly, w, h := centerBox(bw, bh)
	return captureRegion(x0+lx, y0+ly, w, h)
}

func captureHUD() (image.Image, error) {
	x0, y0, bw, bh, ok := overwatchClientBounds()
	if !ok {
		return nil, ErrGameNotVisible
	}
	lx, ly, w, h := hudBox(bw, bh)
	return captureRegion(x0+lx, y0+ly, w, h)
}

func isOverwatchTitle(title string) bool {
	t := strings.TrimSpace(title)
	return t == "Overwatch" || t == "Overwatch 2"
}
