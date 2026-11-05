package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// The downloaded exe finishes its own install: the running version starts it from
// data/update-staging, and it waits for that version to exit, moves itself over the
// old exe and starts the result. Nothing goes through cmd.exe or a .bat, so any
// folder name works, and the old exe stays untouched until the final atomic rename.

const (
	finishFlag      = "--finish-update"
	swapResultFile  = "update-result.txt"
	swapReadyWait   = 30 * time.Second
	swapExitWait    = 30 * time.Second
	swapReplaceWait = 30 * time.Second
)

// startSwap hands over to the staged exe. An error means it did not start, so the
// running version simply keeps working.
func startSwap(exe, staged, dataDir string) error {
	ready := staged + ".ready"
	_ = os.Remove(ready)
	cmd := exec.Command(staged, finishFlag, exe, strconv.Itoa(os.Getpid()), dataDir)
	cmd.Dir = filepath.Dir(exe)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("the new version could not start (blocked by antivirus?): %w", err)
	}
	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(exited)
	}()
	deadline := time.After(swapReadyWait)
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if _, err := os.Stat(ready); err == nil {
				return nil
			}
		case <-exited:
			return fmt.Errorf("the new version closed right after starting (blocked by antivirus?)")
		case <-deadline:
			_ = cmd.Process.Kill()
			select {
			case <-exited:
			case <-time.After(5 * time.Second):
			}
			return fmt.Errorf("the new version did not respond (blocked by antivirus?)")
		}
	}
}

// HandleFinishArgs runs the install step when the process was started by startSwap.
// It returns false for a normal launch.
func HandleFinishArgs(args []string) bool {
	if len(args) != 4 || args[0] != finishFlag {
		return false
	}
	exe, dataDir := args[1], args[3]
	self, err := os.Executable()
	if err == nil {
		_ = os.WriteFile(self+".ready", nil, 0o644)
	}
	if pid, convErr := strconv.Atoi(args[2]); convErr == nil {
		waitForExit(pid, swapExitWait)
	}
	if err == nil {
		err = replaceExe(self, exe)
	}
	if err != nil {
		_ = os.WriteFile(filepath.Join(dataDir, swapResultFile), []byte(failureCode(err)), 0o644)
	}
	// Whichever version is in place now: the app must always come back.
	cmd := exec.Command(exe)
	cmd.Dir = filepath.Dir(exe)
	_ = cmd.Start()
	return true
}

func waitForExit(pid int, timeout time.Duration) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return // already gone
	}
	done := make(chan struct{})
	go func() {
		_, _ = p.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		_ = p.Kill()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}
}

// replaceExe copies src next to dst and renames it over dst, so dst is always either
// the complete old file or the complete new one. The rename is retried because an
// antivirus can hold the exe for a while after its process exits.
func replaceExe(src, dst string) error {
	tmp := dst + ".new"
	if err := copyFile(src, tmp); err != nil {
		return &Failure{CodeWrite, err}
	}
	deadline := time.Now().Add(swapReplaceWait)
	for {
		err := os.Rename(tmp, dst)
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			_ = os.Remove(tmp)
			return &Failure{CodeLocked, err}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// takeSwapResult returns the failure code of the last install and forgets it.
func takeSwapResult(dataDir string) string {
	path := filepath.Join(dataDir, swapResultFile)
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	_ = os.Remove(path)
	return strings.TrimSpace(string(b))
}

// CleanupStaging removes data/update-staging. Run it in the background on startup:
// the installer process that started this one may need a moment to exit.
func CleanupStaging(dataDir string) {
	staging := filepath.Join(dataDir, "update-staging")
	for i := 0; i < 20; i++ {
		if os.RemoveAll(staging) == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}
