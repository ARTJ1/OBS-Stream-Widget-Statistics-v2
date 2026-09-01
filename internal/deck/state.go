package deck

import (
	"strings"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
)

// State is a Stream Dock / AJAZZ-friendly view of current widget stats.
type State struct {
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Rank         int    `json:"rank"`
	RankLabel    string `json:"rankLabel"`
	RankShort    string `json:"rankShort"`
	RankTier     string `json:"rankTier"`
	RankImageURL string `json:"rankImageUrl"`
	Game         string `json:"game"`
	Mode         string `json:"mode"`
	Role         string `json:"role"`
	RoleImageURL string `json:"roleImageUrl"`
	SkinID       string `json:"skinId"`
	WinsColor    string `json:"winsColor,omitempty"`
	LossesColor  string `json:"lossesColor,omitempty"`
}

// FromSnapshot builds deck display fields from a store snapshot.
// baseURL may be empty (relative /overlay/… paths) or e.g. http://127.0.0.1:19123.
func FromSnapshot(snap store.Snapshot, baseURL string) State {
	v := snap.View
	label, short, img, tier := resolveRank(v.Game, v.Rank)
	return State{
		Wins:         v.Wins,
		Losses:       v.Losses,
		Rank:         v.Rank,
		RankLabel:    label,
		RankShort:    short,
		RankTier:     tier,
		RankImageURL: joinURL(baseURL, "/overlay/"+img),
		Game:         v.Game,
		Mode:         v.Mode,
		Role:         v.Role,
		RoleImageURL: joinURL(baseURL, "/overlay/"+roleImagePath(v.Role)),
		SkinID:       snap.Settings.SkinID,
		WinsColor:    snap.Settings.WinsColor,
		LossesColor:  snap.Settings.LossesColor,
	}
}

func joinURL(base, path string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if base == "" {
		return path
	}
	return base + path
}
