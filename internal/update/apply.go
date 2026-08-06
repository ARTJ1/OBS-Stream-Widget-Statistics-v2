package update

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func downloadClient() *http.Client {
	return &http.Client{
		Timeout: 5 * time.Minute,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   20 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   20 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     false,
			MaxIdleConns:          4,
			DisableCompression:    true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 8 {
				return fmt.Errorf("too many redirects")
			}
			req.Header.Set("User-Agent", userAgent)
			return nil
		},
	}
}

// CleanupJunk removes leftover updater files next to the exe (no new folders created).
func CleanupJunk(exeDir string) {
	if exeDir == "" {
		return
	}
	names := []string{
		"widget-stats.exe.new",
		"widget-stats.exe.new.part",
		"widget-stats.exe.part",
		"widget-stats.exe.update.bat",
		"widget_control.lua.new",
		"widget_control.lua.new.part",
		"widget_control.lua.part",
	}
	for _, name := range names {
		_ = os.Remove(filepath.Join(exeDir, name))
	}
	entries, err := os.ReadDir(exeDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			lower := strings.ToLower(name)
			// Odd leftovers like widget_control.lua5 / .lua.new / numbered copies.
			if strings.HasPrefix(lower, "widget_control.lua") && lower != "widget_control.lua" {
				_ = os.Remove(filepath.Join(exeDir, name))
			}
			if strings.HasPrefix(lower, "widget-stats.exe.") &&
				(strings.HasSuffix(lower, ".new") ||
					strings.HasSuffix(lower, ".part") ||
					strings.HasSuffix(lower, ".bat") ||
					strings.Contains(lower, ".update.")) {
				_ = os.Remove(filepath.Join(exeDir, name))
			}
		}
	}
	// Remove auto-created empty (or lua-only) obs/ clutter from older updaters.
	obsDir := filepath.Join(exeDir, "obs")
	if ents, err := os.ReadDir(obsDir); err == nil {
		onlyOurs := true
		for _, e := range ents {
			n := strings.ToLower(e.Name())
			if n != "widget_control.lua" && !strings.HasPrefix(n, "widget_control.lua") {
				onlyOurs = false
				break
			}
		}
		if onlyOurs {
			for _, e := range ents {
				_ = os.Remove(filepath.Join(obsDir, e.Name()))
			}
			_ = os.Remove(obsDir)
		}
	}
}

// Apply downloads the new exe (and optionally updates an existing lua script in place).
// Does not create obs/ or extra lua copies. Restarts only widget-stats.exe.
func Apply(info Info, dataDir string) error {
	if !info.Available {
		return fmt.Errorf("no update available")
	}
	exeURLs := info.ExeURLs()
	if len(exeURLs) == 0 {
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
	CleanupJunk(exeDir)

	staging := filepath.Join(dataDir, "update-staging")
	_ = os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(staging) }()

	client := downloadClient()
	newExe := filepath.Join(staging, "widget-stats.exe")
	if err := downloadWithRetry(client, exeURLs, newExe); err != nil {
		return fmt.Errorf("download exe: %w", err)
	}

	// Lua: only patch paths that already exist; skip if identical.
	if luaURLs := info.LuaURLs(); len(luaURLs) > 0 {
		if targets := existingLuaTargets(exeDir, dataDir); len(targets) > 0 {
			tmpLua := filepath.Join(staging, "widget_control.lua")
			if err := downloadWithRetry(client, luaURLs, tmpLua); err != nil {
				// Lua is optional for most releases — don't fail the whole update.
			} else if b, err := os.ReadFile(tmpLua); err == nil {
				for _, dest := range targets {
					old, _ := os.ReadFile(dest)
					if bytes.Equal(old, b) {
						continue
					}
					_ = os.WriteFile(dest, b, 0o644)
				}
			}
		}
	}

	finalNew := exe + ".new"
	_ = os.Remove(finalNew)
	if err := copyFile(newExe, finalNew); err != nil {
		return fmt.Errorf("stage exe: %w", err)
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-apply is only supported on Windows")
	}
	return scheduleWindowsReplace(exe, finalNew, os.Getpid())
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return closeErr
	}
	return nil
}

func (info Info) ExeURLs() []string {
	return uniqURLs(info.ExeAPIURL, info.ExeURL)
}

func (info Info) LuaURLs() []string {
	return uniqURLs(info.LuaAPIURL, info.LuaURL)
}

func uniqURLs(urls ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

// existingLuaTargets returns only lua files that already exist — never creates folders/copies.
func existingLuaTargets(exeDir, dataDir string) []string {
	seen := map[string]bool{}
	var out []string
	addIfExists := func(p string) {
		p = filepath.Clean(strings.TrimSpace(p))
		if p == "" || seen[p] {
			return
		}
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	addIfExists(filepath.Join(exeDir, "widget_control.lua"))
	if dataDir != "" {
		if b, err := os.ReadFile(filepath.Join(dataDir, "lua_path.txt")); err == nil {
			addIfExists(string(b))
		}
	}
	return out
}

func downloadWithRetry(client *http.Client, urls []string, dest string) error {
	var last error
	for attempt := 0; attempt < 5; attempt++ {
		for _, u := range urls {
			if err := downloadOnce(client, u, dest); err != nil {
				last = err
				_ = os.Remove(dest)
				continue
			}
			return nil
		}
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
	if last == nil {
		return fmt.Errorf("download failed")
	}
	return last
}

func downloadOnce(client *http.Client, url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("Connection", "close")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("%s: %s", res.Status, strings.TrimSpace(string(body)))
	}

	tmp := dest + ".part"
	_ = os.Remove(tmp)
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	written, err := io.Copy(f, io.LimitReader(res.Body, 200<<20))
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	if written < 1024 {
		_ = os.Remove(tmp)
		return fmt.Errorf("file too small (%d bytes)", written)
	}
	if res.ContentLength > 0 && written != res.ContentLength {
		_ = os.Remove(tmp)
		return fmt.Errorf("incomplete download: got %d of %d bytes", written, res.ContentLength)
	}
	_ = os.Remove(dest)
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
