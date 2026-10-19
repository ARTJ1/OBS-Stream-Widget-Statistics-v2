package store

import "testing"

func newTestStore(t *testing.T, mode string) *Store {
	t.Helper()
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetMode(mode); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSetAutoRankRoleModes(t *testing.T) {
	s := newTestStore(t, ModeRolesShared)
	if _, err := s.SetRole(RoleTank); err != nil {
		t.Fatal(err)
	}
	snap, old, applied, err := s.SetAutoRank(RoleDamage, 23)
	if err != nil || !applied || old != 0 {
		t.Fatalf("applied=%v old=%d err=%v", applied, old, err)
	}
	if snap.State.Roles[RoleDamage].Rank != 23 || snap.State.Roles[RoleTank].Rank != 0 {
		t.Fatalf("roles: %+v", snap.State.Roles)
	}
	if snap.State.Role != RoleDamage {
		t.Fatalf("shown role = %s, want damage (the one played)", snap.State.Role)
	}
	if _, _, applied, _ := s.SetAutoRank(AutoRoleOpen, 30); applied {
		t.Fatal("open queue rank applied in a role mode")
	}
}

func TestSetAutoRankRotateKeepsRole(t *testing.T) {
	s := newTestStore(t, ModeRolesRotate)
	before := s.State().Role
	snap, _, applied, _ := s.SetAutoRank(RoleSupport, 10)
	if !applied || snap.State.Roles[RoleSupport].Rank != 10 || snap.State.Role != before {
		t.Fatalf("applied=%v state=%+v", applied, snap.State)
	}
}

func TestSetAutoRankClassic(t *testing.T) {
	s := newTestStore(t, ModeClassic)
	for _, role := range []string{RoleSupport, AutoRoleOpen} {
		snap, _, applied, _ := s.SetAutoRank(role, 17)
		if !applied || snap.State.Rank != 17 {
			t.Fatalf("%s: applied=%v rank=%d", role, applied, snap.State.Rank)
		}
	}
	if _, _, applied, _ := s.SetAutoRank("", 5); applied {
		t.Fatal("rank without a role applied")
	}
}

func TestMoveRoleResult(t *testing.T) {
	s := newTestStore(t, ModeRolesSplit)
	if _, err := s.SetRole(RoleTank); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddWin(); err != nil {
		t.Fatal(err)
	}
	snap, moved, err := s.MoveRoleResult(RoleTank, RoleDamage, "win")
	if err != nil || !moved {
		t.Fatalf("moved=%v err=%v", moved, err)
	}
	if snap.State.Roles[RoleTank].Wins != 0 || snap.State.Roles[RoleDamage].Wins != 1 {
		t.Fatalf("roles: %+v", snap.State.Roles)
	}
	if _, moved, _ := s.MoveRoleResult(RoleTank, RoleDamage, "win"); moved {
		t.Fatal("moved a win that is not there")
	}

	shared := newTestStore(t, ModeRolesShared)
	if _, moved, _ := shared.MoveRoleResult(RoleTank, RoleDamage, "win"); moved {
		t.Fatal("moved a result outside split mode")
	}
}
