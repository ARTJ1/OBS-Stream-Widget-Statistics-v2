package owtracker

import "os"

// DevMode enables test/debug UI and APIs (set WIDGET_STATS_DEV=1 for developers).
func DevMode() bool {
	v := os.Getenv("WIDGET_STATS_DEV")
	return v == "1" || v == "true" || v == "yes"
}

// legacyDetector switches back to the old color/ML detector (dev comparison only):
// WIDGET_STATS_DEV=1 and WIDGET_STATS_LEGACY_DETECTOR=1.
func legacyDetector() bool {
	return DevMode() && os.Getenv("WIDGET_STATS_LEGACY_DETECTOR") == "1"
}
