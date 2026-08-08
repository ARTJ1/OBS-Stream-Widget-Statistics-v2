package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const MaxRankOverwatch = 40 + 500 // Bronze..Champion (40) + Top 500 = 540
const MaxRankApex = 7*4 + 750     // Rookie..Master (28) + Predator 750 = 778

// Deprecated alias — prefer maxRankFor(game).
const MaxRankIndex = MaxRankOverwatch

const (
	GameOverwatch = "overwatch"
	GameApex      = "apex"

	ModeClassic     = "classic"
	ModeRolesShared = "roles_shared" // shared W/L, per-role rank
	ModeRolesSplit  = "roles_split"  // per-role W/L and rank
	ModeRolesRotate = "roles_rotate" // shared W/L, per-role rank, rotate on appear

	RoleTank    = "tank"
	RoleSupport = "support"
	RoleDamage  = "damage"
)

var modeOrder = []string{ModeClassic, ModeRolesShared, ModeRolesSplit, ModeRolesRotate}
var roleOrder = []string{RoleTank, RoleSupport, RoleDamage}
var gameOrder = []string{GameOverwatch, GameApex}

type RoleStats struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Rank   int `json:"rank"`
}

// GameBlob is a snapshot of play stats for one game title.
type GameBlob struct {
	Mode      string               `json:"mode"`
	Role      string               `json:"role"`
	RoleCycle []string             `json:"roleCycle"`
	Wins      int                  `json:"wins"`
	Losses    int                  `json:"losses"`
	Rank      int                  `json:"rank"`
	Roles     map[string]RoleStats `json:"roles"`
}

type State struct {
	Game      string               `json:"game"`
	Mode      string               `json:"mode"`
	Role      string               `json:"role"`
	RoleCycle []string             `json:"roleCycle"`
	Wins      int                  `json:"wins"`
	Losses    int                  `json:"losses"`
	Rank      int                  `json:"rank"`
	Roles     map[string]RoleStats `json:"roles"`
	Saved     map[string]GameBlob  `json:"saved,omitempty"`
}

// View is the currently displayed W/L/rank for overlay clients.
type View struct {
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
	Rank   int    `json:"rank"`
	Mode   string `json:"mode"`
	Role   string `json:"role"`
	Game   string `json:"game"`
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

	IconColor      string `json:"iconColor"`
	VesselStyle    string `json:"vesselStyle"` // classic|royal|rage|inferno|toxic|nova|fang|hard|grim|solid|sharp|award|medal
	SeparatorColor string `json:"separatorColor"`
	AppearEffect   string `json:"appearEffect"` // slide | fade | bounce | zoom
	FillStyle      string `json:"fillStyle"`    // liquid | solid | glow | bubble
	RankFx         string `json:"rankFx"`

	FillLimit       int    `json:"fillLimit"`
	FillDurationMs  int    `json:"fillDurationMs"`
	EmptyDurationMs int    `json:"emptyDurationMs"`
	EmptyEffect     string `json:"emptyEffect"`
	NumberAnimMs    int    `json:"numberAnimMs"`
	VesselWave      bool   `json:"vesselWave"`
	IdlePulse       bool   `json:"idlePulse"`
	SkinID          string `json:"skinId"`
	UiLang          string `json:"uiLang"`  // ru | en
	UiTheme         string `json:"uiTheme"` // dark | light

	ObsHost       string `json:"obsHost"`
	ObsPort       int    `json:"obsPort"`
	ObsPassword   string `json:"obsPassword"`
	ObsScene      string `json:"obsScene"`
	ObsSourceName string `json:"obsSourceName"`
}

type Snapshot struct {
	State    State    `json:"state"`
	Settings Settings `json:"settings"`
	View     View     `json:"view"`
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
		VesselStyle:     "classic",
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
		UiLang:          "ru",
		UiTheme:         "dark",
		ObsHost:         "127.0.0.1",
		ObsPort:         4455,
		ObsPassword:     "",
		ObsScene:        "",
		ObsSourceName:   "Widget Stats v2",
	}
}

func defaultRoles() map[string]RoleStats {
	return map[string]RoleStats{
		RoleTank:    {},
		RoleSupport: {},
		RoleDamage:  {},
	}
}

func DefaultState() State {
	return State{
		Game:      GameOverwatch,
		Mode:      ModeClassic,
		Role:      RoleTank,
		RoleCycle: append([]string(nil), roleOrder...),
		Roles:     defaultRoles(),
		Saved:     map[string]GameBlob{},
	}
}

func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		dir:      dataDir,
		state:    DefaultState(),
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

func maxRankFor(game string) int {
	if game == GameApex {
		return MaxRankApex
	}
	return MaxRankOverwatch
}

func clampRank(n int) int {
	return clampRankFor(GameOverwatch, n)
}

func clampRankFor(game string, n int) int {
	if n < 0 {
		return 0
	}
	max := maxRankFor(game) - 1
	if n > max {
		return max
	}
	return n
}

func clampNonNeg(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func isValidGame(game string) bool {
	return game == GameOverwatch || game == GameApex
}

func isValidRole(id string) bool {
	return id == RoleTank || id == RoleSupport || id == RoleDamage
}

func isValidMode(mode string) bool {
	return mode == ModeClassic || mode == ModeRolesShared || mode == ModeRolesSplit || mode == ModeRolesRotate
}

func normalizeRoleCycle(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(roleOrder))
	for _, id := range roleOrder {
		for _, cand := range in {
			if cand == id && !seen[id] {
				seen[id] = true
				out = append(out, id)
				break
			}
		}
	}
	if len(out) == 0 {
		return append([]string(nil), roleOrder...)
	}
	return out
}

func (s *Store) clamp() {
	if !isValidGame(s.state.Game) {
		s.state.Game = GameOverwatch
	}
	if s.state.Game == GameApex {
		// Apex has no role ladder — keep classic display behavior.
		s.state.Mode = ModeClassic
	}
	if !isValidMode(s.state.Mode) {
		s.state.Mode = ModeClassic
	}
	if !isValidRole(s.state.Role) {
		s.state.Role = RoleTank
	}
	s.state.RoleCycle = normalizeRoleCycle(s.state.RoleCycle)
	if s.state.Mode == ModeRolesRotate {
		inCycle := false
		for _, id := range s.state.RoleCycle {
			if id == s.state.Role {
				inCycle = true
				break
			}
		}
		if !inCycle {
			s.state.Role = s.state.RoleCycle[0]
		}
	}
	if s.state.Roles == nil {
		s.state.Roles = defaultRoles()
	}
	if s.state.Saved == nil {
		s.state.Saved = map[string]GameBlob{}
	}
	for _, id := range roleOrder {
		rs, ok := s.state.Roles[id]
		if !ok {
			s.state.Roles[id] = RoleStats{}
			continue
		}
		rs.Wins = clampNonNeg(rs.Wins)
		rs.Losses = clampNonNeg(rs.Losses)
		rs.Rank = clampRankFor(s.state.Game, rs.Rank)
		s.state.Roles[id] = rs
	}
	s.state.Wins = clampNonNeg(s.state.Wins)
	s.state.Losses = clampNonNeg(s.state.Losses)
	s.state.Rank = clampRankFor(s.state.Game, s.state.Rank)

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
	if s.settings.VesselStyle == "" {
		s.settings.VesselStyle = "classic"
	}
	switch s.settings.VesselStyle {
	case "classic", "royal", "rage", "inferno", "toxic", "nova",
		"fang", "hard", "grim", "solid", "sharp", "award", "medal":
		// ok
	case "brutal", "emoji":
		s.settings.VesselStyle = "fang"
	case "spike", "simple":
		s.settings.VesselStyle = "hard"
	case "shard", "tech", "free", "cross":
		s.settings.VesselStyle = "grim"
	default:
		// Dropped pairs (bones/combat) and unknown values fall back.
		s.settings.VesselStyle = "classic"
	}
	if s.settings.AppearEffect == "" {
		s.settings.AppearEffect = "slide"
	}
	if s.settings.UiLang != "en" && s.settings.UiLang != "ru" {
		s.settings.UiLang = "ru"
	}
	if s.settings.UiTheme != "light" && s.settings.UiTheme != "dark" {
		s.settings.UiTheme = "dark"
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

func (st State) View() View {
	game := st.Game
	if !isValidGame(game) {
		game = GameOverwatch
	}
	v := View{Mode: st.Mode, Role: st.Role, Game: game}
	if game == GameApex {
		v.Wins, v.Losses, v.Rank = st.Wins, st.Losses, st.Rank
		v.Mode = ModeClassic
		return v
	}
	switch st.Mode {
	case ModeRolesShared, ModeRolesRotate:
		rs := st.Roles[st.Role]
		v.Wins, v.Losses, v.Rank = st.Wins, st.Losses, rs.Rank
	case ModeRolesSplit:
		rs := st.Roles[st.Role]
		v.Wins, v.Losses, v.Rank = rs.Wins, rs.Losses, rs.Rank
	default:
		v.Wins, v.Losses, v.Rank = st.Wins, st.Losses, st.Rank
		v.Mode = ModeClassic
	}
	return v
}

func cloneRoles(in map[string]RoleStats) map[string]RoleStats {
	out := defaultRoles()
	if in == nil {
		return out
	}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (s *Store) captureBlob() GameBlob {
	return GameBlob{
		Mode:      s.state.Mode,
		Role:      s.state.Role,
		RoleCycle: append([]string(nil), s.state.RoleCycle...),
		Wins:      s.state.Wins,
		Losses:    s.state.Losses,
		Rank:      s.state.Rank,
		Roles:     cloneRoles(s.state.Roles),
	}
}

func (s *Store) applyBlob(blob GameBlob) {
	s.state.Mode = blob.Mode
	s.state.Role = blob.Role
	s.state.RoleCycle = append([]string(nil), blob.RoleCycle...)
	s.state.Wins = blob.Wins
	s.state.Losses = blob.Losses
	s.state.Rank = blob.Rank
	s.state.Roles = cloneRoles(blob.Roles)
}

func (s *Store) SetGame(game string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !isValidGame(game) {
		return Snapshot{}, fmt.Errorf("invalid game %q", game)
	}
	cur := s.state.Game
	if !isValidGame(cur) {
		cur = GameOverwatch
	}
	if game == cur {
		return s.snapLocked(), nil
	}
	if s.state.Saved == nil {
		s.state.Saved = map[string]GameBlob{}
	}
	s.state.Saved[cur] = s.captureBlob()
	if next, ok := s.state.Saved[game]; ok {
		s.applyBlob(next)
	} else {
		def := DefaultState()
		def.Game = game
		if game == GameApex {
			def.Mode = ModeClassic
		}
		s.applyBlob(GameBlob{
			Mode:      def.Mode,
			Role:      def.Role,
			RoleCycle: def.RoleCycle,
			Wins:      0,
			Losses:    0,
			Rank:      0,
			Roles:     defaultRoles(),
		})
	}
	s.state.Game = game
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) CycleGame() (Snapshot, error) {
	s.mu.RLock()
	idx := 0
	for i, g := range gameOrder {
		if g == s.state.Game {
			idx = i
			break
		}
	}
	next := gameOrder[(idx+1)%len(gameOrder)]
	s.mu.RUnlock()
	return s.SetGame(next)
}

func (s *Store) snapLocked() Snapshot {
	return Snapshot{State: s.state, Settings: s.settings, View: s.state.View()}
}

func (s *Store) Dir() string {
	return s.dir
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapLocked()
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

func (s *Store) rolePtr() *RoleStats {
	rs := s.state.Roles[s.state.Role]
	s.state.Roles[s.state.Role] = rs
	r := s.state.Roles[s.state.Role]
	return &r
}

func (s *Store) getRole(id string) RoleStats {
	if s.state.Roles == nil {
		return RoleStats{}
	}
	return s.state.Roles[id]
}

func (s *Store) setRole(id string, rs RoleStats) {
	if s.state.Roles == nil {
		s.state.Roles = defaultRoles()
	}
	rs.Wins = clampNonNeg(rs.Wins)
	rs.Losses = clampNonNeg(rs.Losses)
	rs.Rank = clampRankFor(s.state.Game, rs.Rank)
	s.state.Roles[id] = rs
}

func (s *Store) AddWin() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state.Mode {
	case ModeRolesSplit:
		rs := s.getRole(s.state.Role)
		rs.Wins++
		s.setRole(s.state.Role, rs)
	default:
		s.state.Wins++
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) AddLoss() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state.Mode {
	case ModeRolesSplit:
		rs := s.getRole(s.state.Role)
		rs.Losses++
		s.setRole(s.state.Role, rs)
	default:
		s.state.Losses++
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) RankUp() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	max := maxRankFor(s.state.Game) - 1
	switch s.state.Mode {
	case ModeRolesShared, ModeRolesSplit, ModeRolesRotate:
		rs := s.getRole(s.state.Role)
		if rs.Rank < max {
			rs.Rank++
		}
		s.setRole(s.state.Role, rs)
	default:
		if s.state.Rank < max {
			s.state.Rank++
		}
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) RankDown() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state.Mode {
	case ModeRolesShared, ModeRolesSplit, ModeRolesRotate:
		rs := s.getRole(s.state.Role)
		if rs.Rank > 0 {
			rs.Rank--
		}
		s.setRole(s.state.Role, rs)
	default:
		if s.state.Rank > 0 {
			s.state.Rank--
		}
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

// RankUpRole bumps rank for an explicit role (Stream Deck / hotkeys), independent of current selection.
func (s *Store) RankUpRole(role string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !isValidRole(role) {
		return Snapshot{}, fmt.Errorf("invalid role %q", role)
	}
	rs := s.getRole(role)
	if rs.Rank < maxRankFor(s.state.Game)-1 {
		rs.Rank++
	}
	s.setRole(role, rs)
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

// RankDownRole lowers rank for an explicit role (Stream Deck / hotkeys).
func (s *Store) RankDownRole(role string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !isValidRole(role) {
		return Snapshot{}, fmt.Errorf("invalid role %q", role)
	}
	rs := s.getRole(role)
	if rs.Rank > 0 {
		rs.Rank--
	}
	s.setRole(role, rs)
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) Reset() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch s.state.Mode {
	case ModeRolesSplit:
		rs := s.getRole(s.state.Role)
		rs.Wins = 0
		rs.Losses = 0
		s.setRole(s.state.Role, rs)
	default:
		s.state.Wins = 0
		s.state.Losses = 0
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) SetRank(rank int) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rank = clampRankFor(s.state.Game, rank)
	switch s.state.Mode {
	case ModeRolesShared, ModeRolesSplit, ModeRolesRotate:
		rs := s.getRole(s.state.Role)
		rs.Rank = rank
		s.setRole(s.state.Role, rs)
	default:
		s.state.Rank = rank
	}
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) CycleMode() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Game == GameApex {
		return s.snapLocked(), nil
	}
	idx := 0
	for i, m := range modeOrder {
		if m == s.state.Mode {
			idx = i
			break
		}
	}
	s.state.Mode = modeOrder[(idx+1)%len(modeOrder)]
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) CycleRole() (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Game == GameApex {
		return s.snapLocked(), nil
	}
	order := roleOrder
	if s.state.Mode == ModeRolesRotate {
		order = s.state.RoleCycle
		if len(order) == 0 {
			order = roleOrder
		}
	}
	idx := 0
	for i, r := range order {
		if r == s.state.Role {
			idx = i
			break
		}
	}
	s.state.Role = order[(idx+1)%len(order)]
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) SetMode(mode string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Game == GameApex {
		s.state.Mode = ModeClassic
	} else {
		s.state.Mode = mode
	}
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) SetRole(role string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Role = role
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) SetRoleCycle(roles []string) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.RoleCycle = roles
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) SetState(st State) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = st
	s.clamp()
	if err := s.persistState(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}

func (s *Store) UpdateSettings(next Settings) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = next
	s.clamp()
	if err := s.persistSettings(); err != nil {
		return Snapshot{}, err
	}
	return s.snapLocked(), nil
}
