package update

import (
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
			ForceAttemptHTTP2:     false, // GitHub CDN is more reliable on HTTP/1.1 for large binaries
			MaxIdleConns:          4,
			DisableCompression:    true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 8 {
				return fmt.Errorf("too many redirects")
			}
			// Keep UA on redirects to objects.githubusercontent.com
			req.Header.Set("User-Agent", userAgent)
			return nil
		},
	}
}

// Apply downloads release assets and schedules a restart that replaces the running exe.
// OBS is not restarted — only widget-stats.exe.
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
	newExe := exe + ".new"
	_ = os.Remove(newExe)

	client := downloadClient()
	if err := downloadWithRetry(client, exeURLs, newExe); err != nil {
		_ = os.Remove(newExe)
		return fmt.Errorf("download exe: %w", err)
	}

	if luaURLs := info.LuaURLs(); len(luaURLs) > 0 {
		tmpLua := filepath.Join(exeDir, "widget_control.lua.new")
		_ = os.Remove(tmpLua)
		if err := downloadWithRetry(client, luaURLs, tmpLua); err != nil {
			_ = os.Remove(tmpLua)
			return fmt.Errorf("download lua: %w", err)
		}
		b, err := os.ReadFile(tmpLua)
		if err != nil {
			return err
		}
		for _, dest := range luaTargets(exeDir, dataDir) {
			_ = os.MkdirAll(filepath.Dir(dest), 0o755)
			_ = os.WriteFile(dest, b, 0o644)
		}
		_ = os.Remove(tmpLua)
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("auto-apply is only supported on Windows")
	}
	return scheduleWindowsReplace(exe, newExe, os.Getpid())
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
