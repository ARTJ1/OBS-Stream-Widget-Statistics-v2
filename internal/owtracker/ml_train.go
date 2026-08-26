package owtracker

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"time"
)

// SaveTrainingSample copies the last probe crop into data/ml/train/{label}/ for retraining.
func (t *Tracker) SaveTrainingSample(label string) (string, error) {
	switch label {
	case "win", "loss", "none":
	default:
		return "", fmt.Errorf("label must be win, loss, or none")
	}
	img, err := t.lastProbeImage()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(t.dataDir, "ml", "train", label)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%s_%s.png", label, time.Now().Format("20060102_150405"))
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return "", err
	}
	return path, nil
}
