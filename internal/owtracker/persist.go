package owtracker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type savedHashes struct {
	Win            string `json:"win,omitempty"`
	Loss           string `json:"loss,omitempty"`
	WinQuality     int    `json:"winQuality,omitempty"`
	LossQuality    int    `json:"lossQuality,omitempty"`
	MatchThreshold int    `json:"matchThreshold,omitempty"`
	WinCropW       int    `json:"winCropW,omitempty"`
	WinCropH       int    `json:"winCropH,omitempty"`
	LossCropW      int    `json:"lossCropW,omitempty"`
	LossCropH      int    `json:"lossCropH,omitempty"`
	Zone           Zone   `json:"zone,omitempty"`
	CustomWin      bool   `json:"customWin,omitempty"`
	CustomLoss     bool   `json:"customLoss,omitempty"`
	CustomZone     bool   `json:"customZone,omitempty"`
}

func effectiveConfig(saved savedHashes) savedHashes {
	out := savedHashes{
		Win:            defaultWinHash,
		Loss:           defaultLossHash,
		WinQuality:     100,
		LossQuality:    100,
		MatchThreshold: saved.MatchThreshold,
		Zone:           defaultZone,
	}
	if saved.CustomWin && saved.Win != "" {
		out.Win = saved.Win
		out.WinQuality = saved.WinQuality
		out.WinCropW = saved.WinCropW
		out.WinCropH = saved.WinCropH
		out.CustomWin = true
	}
	if saved.CustomLoss && saved.Loss != "" {
		out.Loss = saved.Loss
		out.LossQuality = saved.LossQuality
		out.LossCropW = saved.LossCropW
		out.LossCropH = saved.LossCropH
		out.CustomLoss = true
	}
	if saved.CustomZone && saved.Zone.Valid() {
		out.Zone = saved.Zone
		out.CustomZone = true
	}
	if out.MatchThreshold < 5 || out.MatchThreshold > 32 {
		out.MatchThreshold = defaultMatchThreshold
	}
	return out
}

func templatesPath(dir string) string {
	return filepath.Join(dir, "ow_phash.json")
}

func loadSavedHashes(dir string) savedHashes {
	var out savedHashes
	b, err := os.ReadFile(templatesPath(dir))
	if err != nil {
		return out
	}
	_ = json.Unmarshal(b, &out)
	return out
}

func saveHashes(dir string, h savedHashes) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(templatesPath(dir), b, 0o644)
}

type hashStore struct {
	mu   sync.RWMutex
	win  string
	loss string
}

func (s *hashStore) get() (win, loss string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.win, s.loss
}

func (s *hashStore) setWin(v string) {
	s.mu.Lock()
	s.win = v
	s.mu.Unlock()
}

func (s *hashStore) setLoss(v string) {
	s.mu.Lock()
	s.loss = v
	s.mu.Unlock()
}
