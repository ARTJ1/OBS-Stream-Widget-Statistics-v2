package store

import (
	"reflect"
	"testing"
)

func TestRecentResults(t *testing.T) {
	s := newTestStore(t, ModeClassic)
	for _, f := range []func() (Snapshot, error){s.AddWin, s.AddLoss, s.AddWin, s.AddWin} {
		if _, err := f(); err != nil {
			t.Fatal(err)
		}
	}
	// Undo of a mistaken win drops the latest "w", not the oldest.
	snap, _ := s.SubWin()
	if want := []string{"w", "l", "w"}; !reflect.DeepEqual(snap.View.Recent, want) {
		t.Fatalf("recent = %v, want %v", snap.View.Recent, want)
	}
	for i := 0; i < 12; i++ {
		snap, _ = s.AddLoss()
	}
	if len(snap.View.Recent) != maxRecent {
		t.Fatalf("recent len = %d, want %d", len(snap.View.Recent), maxRecent)
	}
	snap, _ = s.Reset()
	if len(snap.View.Recent) != 0 {
		t.Fatalf("recent after reset = %v", snap.View.Recent)
	}
}
