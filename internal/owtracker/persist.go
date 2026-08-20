package owtracker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type savedHashes struct {
	Win         string `json:"win"`
	Loss        string `json:"loss"`
	WinQuality  int    `json:"winQuality,omitempty"`
	LossQuality int    `json:"lossQuality,omitempty"`
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
