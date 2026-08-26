//go:build windows

package owtracker

// OCR is not used in production (text net handles banner recognition).
func extractEmbeddedTesseract(dataDir string) (exePath string, ok bool, err error) {
	return "", false, nil
}
