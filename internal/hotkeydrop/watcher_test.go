package hotkeydrop

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func drop(t *testing.T, dir, name, body string, mtime time.Time) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	return p
}

func collect(dir string, since, now time.Time) []Action {
	var got []Action
	drain(dir, since, now, func(a Action) { got = append(got, a) })
	return got
}

// A "quit" left over from closing OBS while the server was down must not kill
// the next server start.
func TestDrainDiscardsDropsFromBeforeStart(t *testing.T) {
	dir := t.TempDir()
	start := time.Now()
	stale := drop(t, dir, "hk_0000000001_000001_1234.wreq", "quit", start.Add(-time.Minute))
	drop(t, dir, "hk_0000000002_000001_1234.wreq", "win", start.Add(time.Second))

	got := collect(dir, start, start.Add(2*time.Second))
	if len(got) != 1 || got[0] != Win {
		t.Fatalf("got %v, want [win]", got)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale drop not removed: %v", err)
	}
}

func TestDrainKeepsFreshEmptyDropForRetry(t *testing.T) {
	dir := t.TempDir()
	start := time.Now()
	p := drop(t, dir, "hk_0000000001_000001_1234.wreq", "", start.Add(time.Second))

	if got := collect(dir, start, start.Add(time.Second+100*time.Millisecond)); len(got) != 0 {
		t.Fatalf("got %v from empty drop", got)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("fresh empty drop was removed before the writer finished: %v", err)
	}

	// The writer finishes; the next tick applies it.
	if err := os.WriteFile(p, []byte("loss"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := collect(dir, start, time.Now()); len(got) != 1 || got[0] != Loss {
		t.Fatalf("got %v, want [loss]", got)
	}
}

func TestDrainRemovesOldEmptyDrop(t *testing.T) {
	dir := t.TempDir()
	start := time.Now()
	p := drop(t, dir, "hk_0000000001_000001_1234.wreq", "", start.Add(time.Second))
	collect(dir, start, start.Add(time.Second+emptyGrace+time.Millisecond))
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("old empty drop not removed: %v", err)
	}
}

func TestDrainAppliesInPressOrderAndIgnoresTmp(t *testing.T) {
	dir := t.TempDir()
	start := time.Now()
	at := start.Add(time.Second)
	// Same second: sequence 10 must come after 9 (zero-padded names).
	drop(t, dir, "hk_0000000005_000010_1111.wreq", "win_down", at)
	drop(t, dir, "hk_0000000005_000009_2222.wreq", "win", at)
	drop(t, dir, "hk_0000000005_000011_3333.tmp", "reset", at)

	got := collect(dir, start, at)
	if len(got) != 2 || got[0] != Win || got[1] != WinDown {
		t.Fatalf("got %v, want [win win_down]", got)
	}
}
