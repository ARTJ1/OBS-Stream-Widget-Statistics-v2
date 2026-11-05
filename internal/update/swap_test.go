//go:build windows

package update

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const swapTestEnv = "WIDGET_SWAP_TEST"

// The test binary plays all three roles: old version, installer and new version.
func TestMain(m *testing.M) {
	if HandleFinishArgs(os.Args[1:]) {
		return
	}
	if os.Getenv(swapTestEnv) != "" {
		swapTestProcess()
		return
	}
	os.Exit(m.Run())
}

func swapTestProcess() {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	started := filepath.Join(dir, "started")
	if _, err := os.Stat(started); err == nil {
		_ = os.WriteFile(filepath.Join(dir, "ran-new"), nil, 0o644)
		return
	}
	_ = os.WriteFile(started, nil, 0o644)
	data := filepath.Join(dir, "data")
	staged := filepath.Join(data, "update-staging", "widget-stats.exe")
	if err := startSwap(exe, staged, data); err != nil {
		_ = os.WriteFile(filepath.Join(dir, "swap-error"), []byte(err.Error()), 0o644)
	}
}

func TestSwapWorksInAnyFolder(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	image, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	// Bytes after the image keep it runnable and mark the file as the new version.
	newImage := append(append([]byte{}, image...), "new version"...)

	for _, name := range []string{"ascii", "with space", "Стрим", "Мой виджет", "R&D", "R & D", "100%", "Widget (1)", "Wow!"} {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), name)
			staging := filepath.Join(dir, "data", "update-staging")
			if err := os.MkdirAll(staging, 0o755); err != nil {
				t.Fatal(err)
			}
			exe := filepath.Join(dir, "widget-stats.exe")
			if err := os.WriteFile(exe, image, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(staging, "widget-stats.exe"), newImage, 0o755); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(exe)
			cmd.Env = append(os.Environ(), swapTestEnv+"=1")
			if err := cmd.Run(); err != nil {
				t.Fatal(err)
			}
			if !waitForFile(filepath.Join(dir, "ran-new"), 30*time.Second) {
				msg, _ := os.ReadFile(filepath.Join(dir, "swap-error"))
				t.Fatalf("new version did not start; swap error: %q", msg)
			}
			got, err := os.ReadFile(exe)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, newImage) {
				t.Fatal("exe was not replaced by the new version")
			}
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				switch e.Name() {
				case "widget-stats.exe", "data", "started", "ran-new":
				default:
					t.Errorf("leftover next to the exe: %s", e.Name())
				}
			}
		})
	}
}

func TestStartSwapFailsCleanlyOnBrokenFile(t *testing.T) {
	dir := t.TempDir()
	staged := filepath.Join(dir, "broken.exe")
	if err := os.WriteFile(staged, []byte("not an executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(dir, "widget-stats.exe")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := startSwap(exe, staged, dir); err == nil {
		t.Fatal("expected an error")
	}
	if got, _ := os.ReadFile(exe); string(got) != "old" {
		t.Fatalf("exe changed: %q", got)
	}
}

func TestSwapResultIsReadOnce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, swapResultFile), []byte("locked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := takeSwapResult(dir); got != "locked" {
		t.Fatalf("got %q", got)
	}
	if got := takeSwapResult(dir); got != "" {
		t.Fatalf("second read: %q", got)
	}
}

func waitForFile(path string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
