package runtimeinfo

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultPort = 19123

type Info struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	BaseURL    string `json:"baseUrl"`
	OverlayURL string `json:"overlayUrl"`
	AdminURL   string `json:"adminUrl"`
	Version    string `json:"version,omitempty"`
	PID        int    `json:"pid"`
}

func FindPort(start int) (int, net.Listener, error) {
	for port := start; port < start+50; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return port, ln, nil
		}
	}
	return 0, nil, fmt.Errorf("no free port near %d", start)
}

// OverlayURLFor builds a cache-busted overlay URL so OBS Browser Source reloads after updates.
func OverlayURLFor(base, ver string) string {
	base = strings.TrimRight(base, "/")
	ver = strings.TrimSpace(ver)
	if ver == "" {
		ver = "dev"
	}
	return fmt.Sprintf("%s/overlay/?v=%s", base, url.QueryEscape(ver))
}

func Build(host string, port int, ver string) Info {
	base := fmt.Sprintf("http://%s:%d", host, port)
	return Info{
		Host:       host,
		Port:       port,
		BaseURL:    base,
		OverlayURL: OverlayURLFor(base, ver),
		AdminURL:   base + "/admin/",
		Version:    ver,
		PID:        os.Getpid(),
	}
}

func Save(dataDir string, info Info) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "runtime.json"), b, 0o644)
}

func LockPath(dataDir string) string {
	return filepath.Join(dataDir, "widget-stats.lock")
}

func AcquireLock(dataDir string) (*os.File, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(LockPath(dataDir), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f); err != nil {
		_ = f.Close()
		return nil, err
	}
	_ = f.Truncate(0)
	_, _ = f.Seek(0, 0)
	_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())
	return f, nil
}

// AcquireLockWait retries AcquireLock for up to wait. After an update the new
// exe can start while the old one is still shutting down; waiting a moment
// keeps the restart from failing (and the admin page waiting forever).
func AcquireLockWait(dataDir string, wait time.Duration) (*os.File, error) {
	deadline := time.Now().Add(wait)
	for {
		f, err := AcquireLock(dataDir)
		if err == nil || time.Now().After(deadline) {
			return f, err
		}
		time.Sleep(300 * time.Millisecond)
	}
}

// ListenPreferred waits up to wait for the preferred port (so an open admin page
// reconnects to the same address after a restart), then falls back to FindPort.
func ListenPreferred(port int, wait time.Duration) (int, net.Listener, error) {
	deadline := time.Now().Add(wait)
	for {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return port, ln, nil
		}
		if time.Now().After(deadline) {
			return FindPort(port)
		}
		time.Sleep(300 * time.Millisecond)
	}
}
