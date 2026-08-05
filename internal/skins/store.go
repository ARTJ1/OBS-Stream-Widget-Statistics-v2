package skins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
)

type CustomSkin struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Created  int64          `json:"created"`
	Settings store.Settings `json:"settings"`
}

type Store struct {
	mu   sync.RWMutex
	dir  string
	list []CustomSkin
}

func New(dataDir string) (*Store, error) {
	s := &Store{dir: dataDir, list: []CustomSkin{}}
	_ = os.MkdirAll(dataDir, 0o755)
	_ = s.load()
	return s, nil
}

func (s *Store) path() string {
	return filepath.Join(s.dir, "custom_skins.json")
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path())
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.list)
}

func (s *Store) save() error {
	b, err := json.MarshalIndent(s.list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), b, 0o644)
}

func (s *Store) List() []CustomSkin {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CustomSkin, len(s.list))
	copy(out, s.list)
	return out
}

func (s *Store) Save(name string, settings store.Settings) (CustomSkin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	skin := CustomSkin{
		ID:       "custom_" + time.Now().Format("20060102_150405"),
		Name:     name,
		Created:  time.Now().Unix(),
		Settings: settings,
	}
	skin.Settings.SkinID = skin.ID
	s.list = append(s.list, skin)
	if err := s.save(); err != nil {
		return CustomSkin{}, err
	}
	return skin, nil
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.list[:0]
	found := false
	for _, item := range s.list {
		if item.ID == id {
			found = true
			continue
		}
		next = append(next, item)
	}
	s.list = next
	if found {
		_ = s.save()
	}
	return found
}
