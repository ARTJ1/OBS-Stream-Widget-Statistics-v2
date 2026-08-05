package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const MaxRankIndex = 40 + 500 // Bronze..Champion (40) + Top 500 = 540 entries, indices 0..539

type State struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Rank   int `json:"rank"`
}

type Settings struct {
	BgType          string `json:"bgType"`
	BgColor         string `json:"bgColor"`
	BgImage         string `json:"bgImage"`
	Font            string `json:"font"`
	FontSize        int    `json:"fontSize"`
	WinsColor       string `json:"winsColor"`
	LossesColor     string `json:"lossesColor"`
	RankTextColor   string `json:"rankTextColor"`
	AnimDirection   string `json:"animDirection"`
	AnimDurationIn  int    `json:"animDurationIn"`
	AnimStayTime    int    `json:"animStayTime"`
	AnimDurationOut int    `json:"animDurationOut"`
	HiddenTime      int    `json:"hiddenTime"`

	IconColor       string `json:"iconColor"`
	SeparatorColor  string `json:"separatorColor"`
	AppearEffect    string `json:"appearEffect"` // slide | fade | bounce | zoom
	FillStyle       string `json:"fillStyle"`    // liquid | solid | glow | bubble
	RankFx          string `json:"rankFx"`       // classic | blaze | frost | neon | divine | melt | arcane | toxic | paper

	// Vessel / fill animations
	FillLimit         int    `json:"fillLimit"`
	FillDurationMs    int    `json:"fillDurationMs"`
	EmptyDurationMs   int    `json:"emptyDurationMs"`
	EmptyEffect       string `json:"emptyEffect"` // drain | splash | burst | pour | fade
	NumberAnimMs      int    `json:"numberAnimMs"`
	VesselWave        bool   `json:"vesselWave"`
	IdlePulse         bool   `json:"idlePulse"`
	SkinID            string `json:"skinId"`

	ObsHost       string `json:"obsHost"`
	ObsPort       int    `json:"obsPort"`
	ObsPassword   string `json:"obsPassword"`
	ObsScene      string `json:"obsScene"`
	ObsSourceName string `json:"obsSourceName"`
}

type Snapshot struct {
	State    State    `json:"state"`
	Settings Settings `json:"settings"`
}

type Store struct {
	mu       sync.RWMutex
	dir      string
	state    State
	settings Settings
}

func DefaultSettings() Settings {
	return Settings{
		BgType:          "color",
		BgColor:         "rgba(0,0,0,0.7)",
		BgImage:         "",
		Font:            "Arial, sans-serif",
		FontSize:        16,
		WinsColor:       "#00ff00",
		LossesColor:     "#ff0000",
		RankTextColor:   "#ffffff",
		IconColor:       "#ffffff",
		SeparatorColor:  "rgba(255,255,255,0.55)",
		AppearEffect:    "slide",
		FillStyle:       "liquid",
		RankFx:          "classic",
		AnimDirection:   "left",
		AnimDurationIn:  500,
		AnimStayTime:    10000,
		AnimDurationOut: 500,
		HiddenTime:      8000,
		FillLimit:       10,
		FillDurationMs:  650,
		EmptyDurationMs: 1600,
		EmptyEffect:     "drain",
		NumberAnimMs:    500,
		VesselWave:      true,
		IdlePulse:       true,
		SkinID:          "default",
		ObsHost:         "127.0.0.1",
		ObsPort:         4455,
		ObsPassword:     "",
		ObsScene:        "",
		ObsSourceName:   "Widget Stats v2",
	}
}

func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		dir:      dataDir,
		state:    State{},
		settings: DefaultSettings(),
	}
	_ = s.loadJSON("state.json", &s.state)
	_ = s.loadJSON("settings.json", &s.settings)
	s.clamp()
	return s, nil
}

func (s *Store) loadJSON(name string, dest any) error {
	b, err := os.ReadFile(filepath.Join(s.dir, name))
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func (s *Store) saveJSON(name string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, name), b, 0o644)
}

func (s *Store) clamp() {
	if s.state.Wins < 0 {
		s.state.Wins = 0
	}
	if s.state.Losses < 0 {
		s.state.Losses = 0
	}
	if s.state.Rank < 0 {
		s.state.Rank = 0
	}
	if s.state.Rank > MaxRankIndex-1 {
		s.state.Rank = MaxRankIndex - 1
	}
	if s.settings.FontSize <= 0 {
		s.settings.FontSize = 16
	}
	if s.settings.AnimDirection == "" {
		s.settings.AnimDirection = "left"
	}
	if s.settings.ObsHost == "" {
		s.settings.ObsHost = "127.0.0.1"
	}
	if s.settings.ObsPort <= 0 {
		s.settings.ObsPort = 4455
	}
	if s.settings.ObsSourceName == "" {
		s.settings.ObsSourceName = "Widget Stats v2"
	}
	if s.settings.RankFx == "" {
		s.settings.RankFx = "classic"
	}
	if s.settings.FillStyle == "" {
		s.settings.FillStyle = "liquid"
	}
	if s.settings.AppearEffect == "" {
		s.settings.AppearEffect = "slide"
	}
	if s.settings.FillLimit <= 0 {
		s.settings.FillLimit = 10
	}
	if s.settings.FillDurationMs <= 0 {
		s.settings.FillDurationMs = 450
	}
	if s.settings.EmptyDurationMs <= 0 {
		s.settings.EmptyDurationMs = 1800
	}
	if s.settings.NumberAnimMs <= 0 {
		s.settings.NumberAnimMs = 500
	}
	if s.settings.EmptyEffect == "" {
		s.settings.EmptyEffect = "drain"
	}
}

func (s *Store) Dir() string {
	return s.dir
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{State: s.state, Settings: s.settings}
}

func (s *Store) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Store) Settings() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *Store) persistState() error {
	return s.saveJSON("state.json", s.state)
}

func (s *Store) persistSettings() error {
	return s.saveJSON("settings.json", s.settings)
}

func (s *Store) AddWin() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Wins++
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}

func (s *Store) AddLoss() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Losses++
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}

func (s *Store) RankUp() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Rank < MaxRankIndex-1 {
		s.state.Rank++
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}

func (s *Store) RankDown() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Rank > 0 {
		s.state.Rank--
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}

func (s *Store) Reset() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Wins = 0
	s.state.Losses = 0
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}

func (s *Store) SetRank(rank int) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Rank = rank
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}

func (s *Store) SetState(st State) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = st
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}


func (s *Store) UpdateSettings(next Settings) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = next
	s.clamp()
	if err := s.persistSettings(); err != nil {
		return Snapshot{}, err
	}
	return Snapshot{State: s.state, Settings: s.settings}, nil
}
