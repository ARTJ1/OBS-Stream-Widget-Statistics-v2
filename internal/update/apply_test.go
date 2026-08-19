package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLuaUpdateTargets(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	if err := os.MkdirAll(data, 0o755); err != nil {
		t.Fatal(err)
	}
	obsPath := filepath.Join(dir, "obs", "widget_control.lua")
	scriptPath := `C:\OBS\scripts\widget_control.lua`
	if err := os.WriteFile(filepath.Join(data, "lua_path.txt"), []byte(scriptPath+"\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	targets := luaUpdateTargets(dir, data)
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d: %v", len(targets), targets)
	}
	if targets[0] != filepath.Join(dir, "widget_control.lua") {
		t.Fatalf("unexpected root target: %s", targets[0])
	}
	if targets[1] != obsPath {
		t.Fatalf("unexpected obs target: %s", targets[1])
	}
	if targets[2] != scriptPath {
		t.Fatalf("unexpected lua_path target: %s", targets[2])
	}
}

func TestApplyLuaFileCreatesMissing(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.lua")
	content := []byte("-- test lua\n" + strings.Repeat("x", 300))
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "obs", "widget_control.lua")
	n, err := applyLuaFile(src, []string{dest})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 update, got %d", n)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytesEqual(got, content) {
		t.Fatal("dest content mismatch")
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
