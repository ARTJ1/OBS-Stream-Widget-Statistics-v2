package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Apply downloads release assets and schedules a restart that replaces the running exe.
// OBS is not restarted — only widget-stats.exe.
func Apply(info Info, dataDir string) error {
	if !info.Available {
		return fmt.Errorf("no update available")
	}
	if info.ExeURL == "" {
		return fmt.Errorf("release has no widget-stats.exe asset")
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exe)
	newExe := exe + ".new"

	client := &http.Client{Timeout: 3 * time.Minute}
	if err := download(client, info.ExeURL, newExe); err != nil {
		return fmt.Errorf("download exe: %w", err)
	}

	if info.LuaURL != "" {
		luaTargets := luaTargets(exeDir, dataDir)
		tmpLua := filepath.Join(exeDir, "widget_control.lua.new")
		if err := download(client, info.LuaURL, tmpLua); err != nil {
			return fmt.Errorf("download lua: %w", err)
		}
		b, err := os.ReadFile(tmpLua)
		if err != nil {
			return err
		}
		for _, dest := range luaTargets {
			_ = os.MkdirAll(filepath.Dir(dest), 0o755)
			if err := os.WriteFile(dest, b, 0o644); err != nil {
				// keep going — at least one target may succeed
				continue
			}
		}
		_ = os.Remove(tmpLua)
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-apply is only supported on Windows")
	}
	return scheduleWindowsReplace(exe, newExe, os.Getpid())
}

func luaTargets(exeDir, dataDir string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		p = filepath.Clean(p)
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	add(filepath.Join(exeDir, "widget_control.lua"))
	add(filepath.Join(exeDir, "obs", "widget_control.lua"))
	if dataDir != "" {
		if b, err := os.ReadFile(filepath.Join(dataDir, "lua_path.txt")); err == nil {
			p := strings.TrimSpace(string(b))
			if p != "" {
				add(p)
			}
		}
	}
	return out
}

func download(client *http.Client, url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, res.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, io.LimitReader(res.Body, 200<<20))
	return err
}

func scheduleWindowsReplace(exe, newExe string, pid int) error {
	bat := exe + ".update.bat"
	// Wait for this PID to exit, then replace and restart.
	content := fmt.Sprintf(`@echo off
setlocal
set "EXE=%s"
set "NEW=%s"
set "PID=%d"
:wait
tasklist /FI "PID eq %%PID%%" 2>NUL | find "%%PID%%" >NUL
if not errorlevel 1 (
  timeout /t 1 /nobreak >NUL
  goto wait
)
move /Y "%%NEW%%" "%%EXE%%" >NUL
start "" "%%EXE%%"
del "%%~f0"
`, exe, newExe, pid)
	if err := os.WriteFile(bat, []byte(content), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("cmd.exe", "/C", bat)
	cmd.Dir = filepath.Dir(exe)
	return cmd.Start()
}
