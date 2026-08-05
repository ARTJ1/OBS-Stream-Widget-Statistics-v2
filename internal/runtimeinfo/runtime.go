package runtimeinfo

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const DefaultPort = 19123

type Info struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	BaseURL    string `json:"baseUrl"`
	OverlayURL string `json:"overlayUrl"`
	AdminURL   string `json:"adminUrl"`
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

func Build(host string, port int) Info {
	base := fmt.Sprintf("http://%s:%d", host, port)
	return Info{
		Host:       host,
		Port:       port,
		BaseURL:    base,
		OverlayURL: base + "/overlay/",
		AdminURL:   base + "/admin/",
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
