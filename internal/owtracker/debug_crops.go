package owtracker

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func saveDebugCrop(dataDir, name string, img image.Image) string {
	if img == nil {
		return ""
	}
	dir := filepath.Join(dataDir, "ow_debug_crops")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return ""
	}
	return path
}

func (t *Tracker) templateCropSizes() (winW, winH, lossW, lossH int) {
	s := loadSavedHashes(t.dataDir)
	return s.WinCropW, s.WinCropH, s.LossCropW, s.LossCropH
}
