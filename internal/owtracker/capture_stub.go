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

func isOverwatchForeground() bool { return false }

func foregroundClientBounds() (x, y, w, h int, ok bool) {
	return 0, 0, 0, 0, false
}

func foregroundClientBoundsWithTitle() (x, y, w, h int, title string, ok bool) {
	return 0, 0, 0, 0, "", false
}

func captureScreenZone(zone Zone) (image.Image, string, error) {
	return nil, "", fmt.Errorf("owtracker: screen capture is only implemented on Windows")
}

func captureWindowZone(zone Zone) (image.Image, string, error) {
	return captureScreenZone(zone)
}

func overwatchClientBounds() (x, y, w, h int, ok bool) {
	return 0, 0, 0, 0, false
}

func overwatchGameBounds() (x, y, w, h int, ok bool) {
	return 0, 0, 0, 0, false
}

func overwatchProcessRunning() bool { return false }
