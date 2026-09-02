package deck

import (
	"testing"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
)

func TestResolveOWBronze5(t *testing.T) {
	label, short, img, tier := resolveRank(store.GameOverwatch, 0)
	if label != "Bronze 5" || tier != "Bronze" || short != "B5" {
		t.Fatalf("got label=%q short=%q tier=%q", label, short, tier)
	}
	if img != "assets/uploads/Bronze.png" {
		t.Fatalf("img=%q", img)
	}
}

func TestResolveOWEmerald(t *testing.T) {
	// 4 tiers * 5 + 0 = Emerald 5 at index 20
	label, short, _, tier := resolveRank(store.GameOverwatch, 20)
	if label != "Emerald 5" || tier != "Emerald" || short != "E5" {
		t.Fatalf("got label=%q short=%q tier=%q", label, short, tier)
	}
}

func TestResolveOWGrandmaster(t *testing.T) {
	// 7*5 = 35 → Grandmaster 5
	_, short, _, tier := resolveRank(store.GameOverwatch, 35)
	if tier != "Grandmaster" || short != "GM5" {
		t.Fatalf("short=%q tier=%q", short, tier)
	}
}

func TestResolveApexPredator(t *testing.T) {
	// 7*4 = 28, then Predator 750 at index 28
	label, short, img, tier := resolveRank(store.GameApex, 28)
	if label != "#750" || tier != "Predator" || short != "#750" {
		t.Fatalf("got label=%q short=%q tier=%q", label, short, tier)
	}
	if img != "assets/apex/Predator.webp" {
		t.Fatalf("img=%q", img)
	}
}

func TestFromSnapshotRolesShared(t *testing.T) {
	st := store.DefaultState()
	st.Mode = store.ModeRolesShared
	st.Role = store.RoleDamage
	st.Wins = 3
	st.Losses = 1
	st.Roles[store.RoleTank] = store.RoleStats{Rank: 0}     // Bronze 5 → B5
	st.Roles[store.RoleSupport] = store.RoleStats{Rank: 20} // Emerald 5 → E5
	st.Roles[store.RoleDamage] = store.RoleStats{Rank: 10}  // Gold 5 → G5
	snap := store.Snapshot{State: st, Settings: store.Settings{SkinID: "x"}, View: st.View()}
	d := FromSnapshot(snap, "http://127.0.0.1:19123")
	if d.RankShort != "G5" {
		t.Fatalf("current rankShort=%q", d.RankShort)
	}
	if d.Roles == nil || d.Roles["support"].RankShort != "E5" || d.Roles["tank"].RankShort != "B5" {
		t.Fatalf("roles=%+v", d.Roles)
	}
	if d.Roles["support"].Wins != 3 || d.Roles["tank"].Wins != 3 {
		t.Fatalf("shared wins expected 3, got %+v", d.Roles)
	}
}

func TestFromSnapshotAbsoluteURL(t *testing.T) {
	snap := store.Snapshot{
		State:    store.DefaultState(),
		Settings: store.Settings{SkinID: "cyber-cyan"},
		View: store.View{
			Wins: 12, Losses: 3, Rank: 0,
			Game: store.GameOverwatch, Mode: store.ModeClassic, Role: store.RoleTank,
		},
	}
	d := FromSnapshot(snap, "http://127.0.0.1:19123")
	if d.Wins != 12 || d.Losses != 3 {
		t.Fatalf("wl=%d/%d", d.Wins, d.Losses)
	}
	if d.RankLabel != "Bronze 5" || d.RankShort != "B5" {
		t.Fatalf("label=%q short=%q", d.RankLabel, d.RankShort)
	}
	want := "http://127.0.0.1:19123/overlay/assets/uploads/Bronze.png"
	if d.RankImageURL != want {
		t.Fatalf("url=%q want %q", d.RankImageURL, want)
	}
	if d.SkinID != "cyber-cyan" {
		t.Fatalf("skin=%q", d.SkinID)
	}
}

func TestLadderLengths(t *testing.T) {
	if len(cachedOW) != 45+500 {
		t.Fatalf("ow len=%d", len(cachedOW))
	}
	if len(cachedApex) != 28+750 {
		t.Fatalf("apex len=%d", len(cachedApex))
	}
}
