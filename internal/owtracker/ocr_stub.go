//go:build !windows

package owtracker

import "fmt"

func initOCR(dataDir string) {
	ocrReady = false
}

func runTesseract(imagePath string) (string, error) {
	return "", fmt.Errorf("ocr unsupported on this platform")
}

func runTesseractPSM(imagePath, psm string) (string, error) {
	return "", fmt.Errorf("ocr unsupported on this platform")
}
