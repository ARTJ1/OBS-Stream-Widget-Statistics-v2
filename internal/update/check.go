package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	GitHubOwner = "ARTJ1"
	GitHubRepo  = "OBS-Stream-Widget-Statistics-v2"
	userAgent   = "OBS-Stream-Widget-Statistics-v2-updater"
)

type Asset struct {
	Name               string `json:"name"`
	URL                string `json:"url"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Info struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	Available  bool   `json:"available"`
	Name       string `json:"name,omitempty"`
	Body       string `json:"body,omitempty"`
	HTMLURL    string `json:"htmlUrl,omitempty"`
	CheckedAt  string `json:"checkedAt"`
	ExeURL     string `json:"exeUrl,omitempty"`
	ExeAPIURL  string `json:"exeApiUrl,omitempty"`
	LuaURL     string `json:"luaUrl,omitempty"`
	LuaAPIURL  string `json:"luaApiUrl,omitempty"`
	Error      string `json:"error,omitempty"`
	Skipped    bool   `json:"skipped,omitempty"`
	SkipReason string `json:"skipReason,omitempty"`
}

type Checker struct {
	mu       sync.RWMutex
	current  string
	dataDir  string
	cached   Info
	have     bool
	client   *http.Client
}

func NewChecker(current, dataDir string) *Checker {
	return &Checker{
		current: current,
		dataDir: dataDir,
		client:  &http.Client{Timeout: 12 * time.Second},
	}
}

func (c *Checker) Current() string {
	return c.current
}

func (c *Checker) Last() (Info, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cached, c.have
}

func (c *Checker) Check(force bool) Info {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !force && c.have && c.cached.CheckedAt != "" {
		if t, err := time.Parse(time.RFC3339, c.cached.CheckedAt); err == nil && time.Since(t) < 30*time.Minute {
			return c.cached
		}
	}

	info := Info{
		Current:   c.current,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	cur := normalizeTag(c.current)
	if cur == "" || cur == "dev" {
		info.Skipped = true
		info.SkipReason = "dev build — update check disabled"
		c.cached, c.have = info, true
		_ = c.writeStatus(info)
		return info
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubOwner, GitHubRepo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		info.Error = err.Error()
		c.cached, c.have = info, true
		_ = c.writeStatus(info)
		return info
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := c.client.Do(req)
	if err != nil {
		info.Error = err.Error()
		c.cached, c.have = info, true
		_ = c.writeStatus(info)
		return info
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("github %s: %s", res.Status, strings.TrimSpace(string(body)))
		c.cached, c.have = info, true
		_ = c.writeStatus(info)
		return info
	}

	var rel struct {
		TagName string  `json:"tag_name"`
		Name    string  `json:"name"`
		Body    string  `json:"body"`
		HTMLURL string  `json:"html_url"`
		Assets  []Asset `json:"assets"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		info.Error = err.Error()
		c.cached, c.have = info, true
		_ = c.writeStatus(info)
		return info
	}

	info.Latest = rel.TagName
	info.Name = rel.Name
	info.Body = rel.Body
	info.HTMLURL = rel.HTMLURL
	info.Available = Newer(rel.TagName, c.current)
	for _, a := range rel.Assets {
		name := strings.ToLower(a.Name)
		switch {
		case name == "widget-stats.exe" || strings.HasSuffix(name, "widget-stats.exe"):
			info.ExeURL = a.BrowserDownloadURL
			info.ExeAPIURL = a.URL
		case name == "widget_control.lua" || strings.HasSuffix(name, "widget_control.lua"):
			info.LuaURL = a.BrowserDownloadURL
			info.LuaAPIURL = a.URL
		}
	}
	c.cached, c.have = info, true
	_ = c.writeStatus(info)
	return info
}

func (c *Checker) writeStatus(info Info) error {
	if c.dataDir == "" {
		return nil
	}
	if err := os.MkdirAll(c.dataDir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dataDir, "update.json"), b, 0o644)
}
