package owtracker

import (
	"image"
	"os"
)

func (t *Tracker) reloadTemplateImages() {
	win, _ := loadTemplateImage(t.dataDir, OutcomeWin)
	loss, _ := loadTemplateImage(t.dataDir, OutcomeLoss)
	t.tplMu.Lock()
	t.winTplImg = win
	t.lossTplImg = loss
	t.tplMu.Unlock()
}

func (t *Tracker) templateImages() (win, loss image.Image) {
	t.tplMu.RLock()
	defer t.tplMu.RUnlock()
	return t.winTplImg, t.lossTplImg
}

func loadTemplateImage(dataDir string, kind Outcome) (image.Image, error) {
	saved := effectiveConfig(normalizeSaved(loadSavedHashes(dataDir)))
	custom := saved.CustomWin
	if kind == OutcomeLoss {
		custom = saved.CustomLoss
	}
	if custom {
		path := templateCropPath(dataDir, kind)
		f, err := os.Open(path)
		if err == nil {
			defer f.Close()
			img, err := decodeImage(f)
			if err == nil {
				return img, nil
			}
		}
	}
	return defaultTemplateImage(kind)
}

func (t *Tracker) scoreCrop(crop image.Image, winHash, lossHash string) (winDist, lossDist, winPix, lossPix int) {
	winDist = bestDistance(crop, winHash)
	lossDist = bestDistance(crop, lossHash)
	winRef, lossRef := t.templateImages()
	if winRef != nil {
		winPix = pixelSimilarityScaled(crop, winRef)
	}
	if lossRef != nil {
		lossPix = pixelSimilarityScaled(crop, lossRef)
	}
	return winDist, lossDist, winPix, lossPix
}
