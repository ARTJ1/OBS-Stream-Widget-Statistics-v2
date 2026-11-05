package update

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

// stallTimeout aborts a download attempt that stopped receiving data.
var stallTimeout = 30 * time.Second

// maxUpdateBytes caps release downloads (safety limit for widget-stats.exe).
const maxUpdateBytes = 512 << 20

func downloadClient() *http.Client {
	return &http.Client{
		Timeout: 20 * time.Minute,
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
	// Remove updater leftovers inside obs/ (widget_control.lua.new / .part / numbered
	// copies). The real widget_control.lua is never touched: users load it in OBS from
	// there, and the repo keeps it in obs/. The folder goes only if it ends up empty.
	obsDir := filepath.Join(exeDir, "obs")
	if ents, err := os.ReadDir(obsDir); err == nil {
		for _, e := range ents {
			n := strings.ToLower(e.Name())
			if strings.HasPrefix(n, "widget_control.lua") && n != "widget_control.lua" {
				_ = os.Remove(filepath.Join(obsDir, e.Name()))
			}
		}
		_ = os.Remove(obsDir) // fails (keeps it) unless empty
	}
}

// Apply downloads the new exe and widget_control.lua from the release.
// Lua is written next to the exe (and obs/ if used), plus any path saved in data/lua_path.txt.
func Apply(info Info, dataDir string, report func(phase string, done, total int64)) error {
	if !info.Available {
		return fmt.Errorf("no update available")
	}
	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-apply is only supported on Windows")
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

	// A folder we can't write to (Program Files) should fail now, not after the app closed.
	probe := exe + ".new"
	if err := os.WriteFile(probe, nil, 0o644); err != nil {
		return &Failure{CodeWrite, err}
	}
	_ = os.Remove(probe)

	staging := filepath.Join(dataDir, "update-staging")
	_ = os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return &Failure{CodeWrite, err}
	}

	client := downloadClient()
	newExe := filepath.Join(staging, "widget-stats.exe")
	onBytes := func(done, total int64) { report(PhaseDownloading, done, total) }
	if err := downloadWithRetry(client, exeURLs, newExe, onBytes); err != nil {
		_ = os.RemoveAll(staging)
		return &Failure{failureCodeOr(err, CodeDownload), fmt.Errorf("download exe: %w", err)}
	}

	// Lua ships with every release — always fetch it when the asset exists. It is
	// applied only after the new exe has started, so a failed start changes nothing.
	tmpLua := ""
	if luaURLs := info.LuaURLs(); len(luaURLs) > 0 {
		p := filepath.Join(staging, "widget_control.lua")
		if err := downloadWithRetry(client, luaURLs, p, nil); err != nil {
			log.Printf("update: widget_control.lua download failed: %v", err)
		} else {
			tmpLua = p
		}
	}

	report(PhaseVerifying, 0, 0)
	if err := startSwap(exe, newExe, dataDir); err != nil {
		CleanupStaging(dataDir)
		return &Failure{CodeStart, err}
	}
	if tmpLua != "" {
		if n, err := applyLuaFile(tmpLua, luaUpdateTargets(exeDir, dataDir)); err != nil {
			log.Printf("update: widget_control.lua apply failed: %v", err)
		} else if n > 0 {
			log.Printf("update: widget_control.lua updated (%d path(s))", n)
		}
	}
	return nil
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

// luaUpdateTargets lists every widget_control.lua path the updater should refresh.
func luaUpdateTargets(exeDir, dataDir string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return // e.g. empty data/lua_path.txt (Clean("") would be ".")
		}
		p = filepath.Clean(p)
		if seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	// Standard install layout from Releases (exe + lua in the same folder).
	add(filepath.Join(exeDir, "widget_control.lua"))
	// Repo-style layout: obs/widget_control.lua next to the exe.
	add(filepath.Join(exeDir, "obs", "widget_control.lua"))
	if dataDir != "" {
		add(readLuaPathFile(filepath.Join(dataDir, "lua_path.txt")))
	}
	return out
}

func readLuaPathFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func applyLuaFile(src string, targets []string) (int, error) {
	b, err := os.ReadFile(src)
	if err != nil {
		return 0, err
	}
	if len(b) < 256 {
		return 0, fmt.Errorf("widget_control.lua too small (%d bytes)", len(b))
	}
	updated := 0
	for _, dest := range targets {
		old, _ := os.ReadFile(dest)
		if bytes.Equal(old, b) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return updated, err
		}
		if err := os.WriteFile(dest, b, 0o644); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func downloadWithRetry(client *http.Client, urls []string, dest string, onBytes func(done, total int64)) error {
	var last error
	for attempt := 0; attempt < 5; attempt++ {
		for _, u := range urls {
			if err := downloadOnce(client, u, dest, onBytes); err != nil {
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

func failureCodeOr(err error, fallback string) string {
	if code := failureCode(err); code != CodeUnknown {
		return code
	}
	return fallback
}

// progressWriter counts bytes for the admin page and keeps the stall watchdog quiet.
type progressWriter struct {
	w        io.Writer
	done     int64
	total    int64
	watchdog *time.Timer
	onBytes  func(done, total int64)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.done += int64(n)
	p.watchdog.Reset(stallTimeout)
	if p.onBytes != nil {
		p.onBytes(p.done, p.total)
	}
	return n, err
}

func downloadOnce(client *http.Client, url, dest string, onBytes func(done, total int64)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stalled atomic.Bool
	watchdog := time.AfterFunc(stallTimeout, func() {
		stalled.Store(true)
		cancel()
	})
	defer watchdog.Stop()
	netErr := func(err error) error {
		if stalled.Load() {
			return &Failure{CodeStalled, fmt.Errorf("no data for %s: %w", stallTimeout, err)}
		}
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("Connection", "close")

	res, err := client.Do(req)
	if err != nil {
		return netErr(err)
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
		return &Failure{CodeWrite, err}
	}

	counter := &progressWriter{w: f, total: res.ContentLength, watchdog: watchdog, onBytes: onBytes}
	written, err := io.Copy(counter, io.LimitReader(res.Body, maxUpdateBytes))
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(tmp)
		return netErr(err)
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
