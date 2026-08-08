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
	mu      sync.RWMutex
	current string
	dataDir string
	cached  Info
	have    bool
	client  *http.Client
}

func NewChecker(current, dataDir string) *Checker {
	return &Checker{
		current: current,
		dataDir: dataDir,
		client:  &http.Client{Timeout: 15 * time.Second},
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

	// Prefer the public HTML "latest" redirect — no api.github.com rate limit.
	tag, htmlURL, err := c.latestTagViaRedirect()
	if err != nil || !isReleaseSemver(tag) {
		// Fallback: GitHub API list, pick newest semver that ships widget-stats.exe.
		tag2, html2, apiErr := c.latestTagViaAPI()
		if apiErr != nil {
			if err != nil {
				info.Error = err.Error()
			} else {
				info.Error = apiErr.Error()
			}
			c.cached, c.have = info, true
			_ = c.writeStatus(info)
			return info
		}
		tag, htmlURL = tag2, html2
	}

	info.Latest = tag
	info.HTMLURL = htmlURL
	info.Name = tag
	info.fillDownloadURLs(tag)
	info.Available = Newer(tag, c.current) && info.ExeURL != ""
	c.cached, c.have = info, true
	_ = c.writeStatus(info)
	return info
}

func (info *Info) fillDownloadURLs(tag string) {
	base := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s", GitHubOwner, GitHubRepo, tag)
	info.ExeURL = base + "/widget-stats.exe"
	info.LuaURL = base + "/widget_control.lua"
	if info.HTMLURL == "" {
		info.HTMLURL = fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", GitHubOwner, GitHubRepo, tag)
	}
}

// latestTagViaRedirect follows github.com/.../releases/latest (302 → /tag/vX.Y.Z).
func (c *Checker) latestTagViaRedirect() (tag, htmlURL string, err error) {
	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", GitHubOwner, GitHubRepo)
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))

	loc := res.Header.Get("Location")
	if loc == "" {
		return "", "", fmt.Errorf("github latest: no redirect (status %s)", res.Status)
	}
	tag = tagFromReleaseURL(loc)
	if tag == "" {
		return "", "", fmt.Errorf("github latest: cannot parse tag from %s", loc)
	}
	if strings.HasPrefix(loc, "http") {
		htmlURL = loc
	} else {
		htmlURL = "https://github.com" + loc
	}
	return tag, htmlURL, nil
}

func tagFromReleaseURL(u string) string {
	// .../releases/tag/v2.2.0  or  /ARTJ1/.../releases/tag/v2.2.0
	const marker = "/releases/tag/"
	i := strings.Index(u, marker)
	if i < 0 {
		return ""
	}
	tag := u[i+len(marker):]
	if j := strings.IndexAny(tag, "?#"); j >= 0 {
		tag = tag[:j]
	}
	return strings.Trim(tag, "/")
}

func isReleaseSemver(tag string) bool {
	return parseSemver(normalizeTag(tag)) != nil
}

// latestTagViaAPI lists releases and picks the newest semver with widget-stats.exe.
// Ignores bonus packs like streamdock-icons-v1.0.0.
func (c *Checker) latestTagViaAPI() (tag, htmlURL string, err error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=30", GitHubOwner, GitHubRepo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")
	if tok := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github %s: %s", res.Status, strings.TrimSpace(string(body)))
	}

	var rels []struct {
		TagName    string  `json:"tag_name"`
		HTMLURL    string  `json:"html_url"`
		Draft      bool    `json:"draft"`
		Prerelease bool    `json:"prerelease"`
		Assets     []Asset `json:"assets"`
	}
	if err := json.Unmarshal(body, &rels); err != nil {
		return "", "", err
	}

	var bestTag, bestHTML string
	for _, rel := range rels {
		if rel.Draft || rel.Prerelease || !isReleaseSemver(rel.TagName) {
			continue
		}
		hasExe := false
		for _, a := range rel.Assets {
			name := strings.ToLower(a.Name)
			if name == "widget-stats.exe" || strings.HasSuffix(name, "widget-stats.exe") {
				hasExe = true
				break
			}
		}
		if !hasExe {
			continue
		}
		if bestTag == "" || Newer(rel.TagName, bestTag) {
			bestTag = rel.TagName
			bestHTML = rel.HTMLURL
		}
	}
	if bestTag == "" {
		return "", "", fmt.Errorf("no semver release with widget-stats.exe")
	}
	return bestTag, bestHTML, nil
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
