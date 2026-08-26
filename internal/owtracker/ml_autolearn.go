package owtracker

import (
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker/ml"
)

const (
	autoLearnMaxPerClass   = 24
	autoLearnHashDup       = 12
	autoLearnRetrainEvery  = 3
	autoLearnRetrainWait   = 10 * time.Minute
)

type mlAutoStore struct {
	dataDir string
	mu      sync.Mutex
	pending int
	last    time.Time
}

func newMLAutoStore(dataDir string) *mlAutoStore {
	return &mlAutoStore{dataDir: dataDir}
}

// LearnFromManual saves the last probe crop when the user presses win/loss (correction or manual score).
func (t *Tracker) LearnFromManual(label string) {
	switch label {
	case "win", "loss":
	default:
		return
	}
	img, err := t.lastProbeImage()
	if err != nil || img == nil {
		return
	}
	if t.autoLearn.save(img, label, true) {
		log.Printf("owtracker: ml learn manual %s", label)
	}
}

func (t *Tracker) learnFromAuto(confirmed ProbeResult, img image.Image) {
	if img == nil || confirmed.Match == "" {
		return
	}
	if confirmed.MatchMethod != "ml" && confirmed.MatchMethod != "text" {
		return
	}
	if confirmed.MatchMethod == "ml" && !mlColorAllows(confirmed.Match, confirmed.GoldPct, confirmed.DefeatPct) {
		return
	}
	if t.autoLearn.save(img, confirmed.Match, false) {
		log.Printf("owtracker: ml learn auto %s", confirmed.Match)
	}
}

func (s *mlAutoStore) save(img image.Image, label string, manual bool) bool {
	if img == nil {
		return false
	}
	h, err := computeTemplateHash(img)
	if err != nil || h == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.dataDir, "ml", "auto", label)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	if dup, _ := autoLearnDuplicate(dir, h); dup {
		return false
	}
	if err := autoLearnEvict(dir, autoLearnMaxPerClass-1); err != nil {
		return false
	}
	name := label + "_" + time.Now().Format("20060102_150405") + ".png"
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return false
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		_ = os.Remove(path)
		return false
	}
	f.Close()
	_ = os.WriteFile(path+".hash", []byte(h), 0o644)

	s.pending++
	if s.pending >= autoLearnRetrainEvery && time.Since(s.last) >= autoLearnRetrainWait {
		s.pending = 0
		s.last = time.Now()
		go s.retrainAsync()
	} else if manual {
		// Manual correction — retrain sooner after user fix.
		go func() {
			time.Sleep(500 * time.Millisecond)
			s.retrainAsync()
		}()
	}
	return true
}

func (s *mlAutoStore) retrainAsync() {
	ok, note, err := ml.Retrain(s.dataDir)
	if err != nil {
		log.Printf("owtracker: ml retrain: %v", err)
		return
	}
	if ok {
		log.Printf("owtracker: ml retrained (%s)", note)
	}
}

func autoLearnDuplicate(dir, hash string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	ha, err := hashFromString(hash)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".png" {
			continue
		}
		pngPath := filepath.Join(dir, e.Name())
		hx := readSidecarHash(pngPath)
		if hx == "" {
			f, err := os.Open(pngPath)
			if err == nil {
				img, err := decodeImage(f)
				f.Close()
				if err == nil {
					if hh, err := computeTemplateHash(img); err == nil {
						hx = hh
					}
				}
			}
		}
		if hx == "" {
			continue
		}
		hb, err := hashFromString(hx)
		if err != nil {
			continue
		}
		d, err := ha.Distance(hb)
		if err == nil && d <= autoLearnHashDup {
			return true, nil
		}
	}
	return false, nil
}

func readSidecarHash(pngPath string) string {
	b, err := os.ReadFile(pngPath + ".hash")
	if err != nil {
		return ""
	}
	return string(b)
}

func autoLearnEvict(dir string, keep int) error {
	type item struct {
		path string
		mod  time.Time
	}
	var items []item
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".png" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, item{filepath.Join(dir, e.Name()), info.ModTime()})
	}
	if len(items) <= keep {
		return nil
	}
	sort.Slice(items, func(i, j int) bool { return items[i].mod.Before(items[j].mod) })
	for _, it := range items[:len(items)-keep] {
		_ = os.Remove(it.path)
		_ = os.Remove(it.path + ".hash")
	}
	return nil
}

func (s *mlAutoStore) counts() (win, loss int) {
	for _, label := range []string{"win", "loss"} {
		dir := filepath.Join(s.dataDir, "ml", "auto", label)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		n := 0
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".png" {
				n++
			}
		}
		switch label {
		case "win":
			win = n
		case "loss":
			loss = n
		}
	}
	return win, loss
}
