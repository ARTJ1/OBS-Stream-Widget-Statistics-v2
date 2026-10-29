package skins

import (
	"testing"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
)

func TestRename(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	skin, err := s.Save("Old", store.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}
	ok, err := s.Rename(skin.ID, "New")
	if err != nil || !ok {
		t.Fatalf("rename ok=%v err=%v", ok, err)
	}
	// Reload from disk: the new name must have been saved.
	again, _ := New(dir)
	if got := again.List()[0].Name; got != "New" {
		t.Fatalf("name after reload = %q", got)
	}
	if ok, _ := s.Rename("missing", "x"); ok {
		t.Fatal("renamed a skin that does not exist")
	}
}
