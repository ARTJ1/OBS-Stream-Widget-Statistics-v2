//go:build !windows

package owtracker

import "image"

func isStreamBrowserTitle(title string) bool { return false }

func isIgnoredTestCaptureTitle(title string) bool { return false }

func captureWindowZoneSmart(zone Zone) (image.Image, string, error) {
	return captureWindowZone(zone)
}
