//go:build !windows

package owtracker

func extractEmbeddedTesseract(dataDir string) (exePath string, ok bool, err error) {
	return "", false, nil
}
