package deck

import (
	"strings"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
)

// RoleState is per-role stats for Stream Dock Live Mode (shared/split/rotate).
type RoleState struct {
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Rank         int    `json:"rank"`
	RankLabel    string `json:"rankLabel"`
	RankShort    string `json:"rankShort"`
	RankTier     string `json:"rankTier"`
	RankImageURL string `json:"rankImageUrl,omitempty"`
}

// State is a Stream Dock / AJAZZ-friendly view of current widget stats.
type State struct {
	Wins         int                  `json:"wins"`
	Losses       int                  `json:"losses"`
	Rank         int                  `json:"rank"`
	RankLabel    string               `json:"rankLabel"`
	RankShort    string               `json:"rankShort"`
	RankTier     string               `json:"rankTier"`
	RankImageURL string               `json:"rankImageUrl"`
	Game         string               `json:"game"`
	Mode         string               `json:"mode"`
	Role         string               `json:"role"`
	RoleImageURL string               `json:"roleImageUrl"`
	SkinID       string               `json:"skinId"`
	WinsColor    string               `json:"winsColor,omitempty"`
	LossesColor  string               `json:"lossesColor,omitempty"`
	Roles        map[string]RoleState `json:"roles,omitempty"`
}

// FromSnapshot builds deck display fields from a store snapshot.
// baseURL may be empty (relative /overlay/… paths) or e.g. http://127.0.0.1:19123.
func FromSnapshot(snap store.Snapshot, baseURL string) State {
	v := snap.View
	label, short, img, tier := resolveRank(v.Game, v.Rank)
	out := State{
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
	out.Roles = rolesFromSnapshot(snap, baseURL)
	return out
}

func rolesFromSnapshot(snap store.Snapshot, baseURL string) map[string]RoleState {
	st := snap.State
	if st.Game == store.GameApex {
		return nil
	}
	switch st.Mode {
	case store.ModeRolesShared, store.ModeRolesSplit, store.ModeRolesRotate:
		// ok
	default:
		return nil
	}
	if st.Roles == nil {
		return nil
	}
	out := make(map[string]RoleState, 3)
	for _, id := range []string{store.RoleTank, store.RoleSupport, store.RoleDamage} {
		rs := st.Roles[id]
		wins, losses, rank := rs.Wins, rs.Losses, rs.Rank
		if st.Mode == store.ModeRolesShared || st.Mode == store.ModeRolesRotate {
			wins, losses = st.Wins, st.Losses
		}
		label, short, img, tier := resolveRank(st.Game, rank)
		out[id] = RoleState{
			Wins:         wins,
			Losses:       losses,
			Rank:         rank,
			RankLabel:    label,
			RankShort:    short,
			RankTier:     tier,
			RankImageURL: joinURL(baseURL, "/overlay/"+img),
		}
	}
	return out
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
