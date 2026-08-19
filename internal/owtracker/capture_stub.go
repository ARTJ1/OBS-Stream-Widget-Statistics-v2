//go:build !windows

package owtracker

import (
	"fmt"
	"image"
)

func initDPI() {}

func activeWindowTitle() string { return "" }

func captureRegion(x, y, w, h int) (image.Image, error) {
	return nil, fmt.Errorf("owtracker: screen capture is only implemented on Windows")
}

func screenSize() (int, int) { return 0, 0 }

func overwatchClientBounds() (x, y, w, h int, ok bool) {
	return 0, 0, 0, 0, false
}

func overwatchGameBounds() (x, y, w, h int, ok bool) {
	return 0, 0, 0, 0, false
}

func overwatchProcessRunning() bool { return false }
