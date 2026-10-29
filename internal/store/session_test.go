package store

import "testing"

func TestSessionStartRank(t *testing.T) {
	s := newTestStore(t, ModeRolesShared)
	if _, err := s.SetRole(RoleSupport); err != nil {
		t.Fatal(err)
	}
	snap, _, _, err := s.SetAutoRank(RoleSupport, 20)
	if err != nil {
		t.Fatal(err)
	}
	if snap.View.Rank != 20 || snap.View.StartRank != 0 {
		t.Fatalf("view = %+v, want rank 20 from start 0", snap.View)
	}
	// Another role keeps its own baseline.
	snap, _, _, _ = s.SetAutoRank(RoleTank, 7)
	if snap.View.Role != RoleTank || snap.View.StartRank != 0 {
		t.Fatalf("tank view = %+v", snap.View)
	}
	// Reset starts a new session at the current ranks.
	snap, err = s.Reset()
	if err != nil {
		t.Fatal(err)
	}
	if snap.View.StartRank != 7 {
		t.Fatalf("after reset start = %d, want 7", snap.View.StartRank)
	}
	snap, _ = s.SetRole(RoleSupport)
	if snap.View.StartRank != 20 {
		t.Fatalf("support start after reset = %d, want 20", snap.View.StartRank)
	}
}

func TestSessionStartRankClassic(t *testing.T) {
	s := newTestStore(t, ModeClassic)
	snap, err := s.SetRank(12)
	if err != nil {
		t.Fatal(err)
	}
	if snap.View.StartRank != 0 || snap.View.Rank != 12 {
		t.Fatalf("view = %+v", snap.View)
	}
}

func TestClampScale(t *testing.T) {
	for in, want := range map[int]int{0: 100, -5: 100, 10: MinScale, 100: 100, 164: 164, 900: MaxScale} {
		if got := ClampScale(in); got != want {
			t.Errorf("ClampScale(%d) = %d, want %d", in, got, want)
		}
	}
}
