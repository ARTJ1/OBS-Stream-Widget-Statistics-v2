package obsbridge

import (
	"fmt"
	"sync"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
)

const DefaultSourceName = "Widget Stats v2"

type Config struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Password   string `json:"password"`
	Scene      string `json:"scene"`
	SourceName string `json:"sourceName"`
}

type Status struct {
	Connected  bool   `json:"connected"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Scene      string `json:"scene"`
	SourceName string `json:"sourceName"`
	Error      string `json:"error,omitempty"`
}

type EnsureResult struct {
	Action     string `json:"action"` // exists | added_existing | created
	Scene      string `json:"scene"`
	SourceName string `json:"sourceName"`
	Message    string `json:"message"`
}

type Bridge struct {
	mu     sync.Mutex
	client *goobs.Client
	cfg    Config
	lastErr string
}

func New() *Bridge {
	return &Bridge{
		cfg: Config{
			Host:       "127.0.0.1",
			Port:       4455,
			SourceName: DefaultSourceName,
		},
	}
}

func (b *Bridge) Configure(cfg Config) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port <= 0 {
		cfg.Port = 4455
	}
	if cfg.SourceName == "" {
		cfg.SourceName = DefaultSourceName
	}
	// Reconnect if endpoint/password changed while connected.
	needReconnect := b.client != nil && (b.cfg.Host != cfg.Host || b.cfg.Port != cfg.Port || b.cfg.Password != cfg.Password)
	b.cfg = cfg
	if needReconnect {
		_ = b.client.Disconnect()
		b.client = nil
	}
}

func (b *Bridge) Config() Config {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.cfg
}

func (b *Bridge) Status() Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	return Status{
		Connected:  b.client != nil,
		Host:       b.cfg.Host,
		Port:       b.cfg.Port,
		Scene:      b.cfg.Scene,
		SourceName: b.cfg.SourceName,
		Error:      b.lastErr,
	}
}

func (b *Bridge) Connect() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.client != nil {
		b.lastErr = ""
		return nil
	}
	addr := fmt.Sprintf("%s:%d", b.cfg.Host, b.cfg.Port)
	opts := []goobs.Option{}
	if b.cfg.Password != "" {
		opts = append(opts, goobs.WithPassword(b.cfg.Password))
	}
	client, err := goobs.New(addr, opts...)
	if err != nil {
		b.lastErr = err.Error()
		return err
	}
	b.client = client
	b.lastErr = ""
	return nil
}

func (b *Bridge) Disconnect() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.client != nil {
		_ = b.client.Disconnect()
		b.client = nil
	}
}

func (b *Bridge) ensureClient() (*goobs.Client, Config, error) {
	if b.client == nil {
		addr := fmt.Sprintf("%s:%d", b.cfg.Host, b.cfg.Port)
		opts := []goobs.Option{}
		if b.cfg.Password != "" {
			opts = append(opts, goobs.WithPassword(b.cfg.Password))
		}
		client, err := goobs.New(addr, opts...)
		if err != nil {
			b.lastErr = err.Error()
			return nil, b.cfg, err
		}
		b.client = client
		b.lastErr = ""
	}
	return b.client, b.cfg, nil
}

func (b *Bridge) Scenes() ([]string, string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	client, _, err := b.ensureClient()
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Scenes.GetSceneList()
	if err != nil {
		b.lastErr = err.Error()
		_ = client.Disconnect()
		b.client = nil
		return nil, "", err
	}
	names := make([]string, 0, len(resp.Scenes))
	for _, sc := range resp.Scenes {
		names = append(names, sc.SceneName)
	}
	return names, resp.CurrentProgramSceneName, nil
}

func (b *Bridge) EnsureOnScene(sceneName, overlayURL string) (*EnsureResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	client, cfg, err := b.ensureClient()
	if err != nil {
		return nil, err
	}
	if sceneName == "" {
		sceneName = cfg.Scene
	}
	if sceneName == "" {
		return nil, fmt.Errorf("scene is not selected")
	}
	sourceName := cfg.SourceName
	if sourceName == "" {
		sourceName = DefaultSourceName
	}

	// Validate scene exists.
	list, err := client.Scenes.GetSceneList()
	if err != nil {
		b.lastErr = err.Error()
		return nil, err
	}
	foundScene := false
	for _, sc := range list.Scenes {
		if sc.SceneName == sceneName {
			foundScene = true
			break
		}
	}
	if !foundScene {
		return nil, fmt.Errorf("scene %q not found in OBS", sceneName)
	}

	items, err := client.SceneItems.GetSceneItemList(sceneitems.NewGetSceneItemListParams().WithSceneName(sceneName))
	if err != nil {
		b.lastErr = err.Error()
		return nil, err
	}
	for _, item := range items.SceneItems {
		if item.SourceName == sourceName {
			// Already on scene — update URL only, never duplicate.
			_, _ = client.Inputs.SetInputSettings(inputs.NewSetInputSettingsParams().
				WithInputName(sourceName).
				WithInputSettings(browserSettings(overlayURL)))
			return &EnsureResult{
				Action:     "exists",
				Scene:      sceneName,
				SourceName: sourceName,
				Message:    "Виджет уже есть на сцене — URL обновлён, дубликат не создан",
			}, nil
		}
	}

	// Source may exist globally (other scene) — add as scene item only.
	inputsList, err := client.Inputs.GetInputList()
	if err != nil {
		b.lastErr = err.Error()
		return nil, err
	}
	existsGlobal := false
	for _, in := range inputsList.Inputs {
		if in.InputName == sourceName {
			existsGlobal = true
			break
		}
	}
	if existsGlobal {
		_, err = client.SceneItems.CreateSceneItem(sceneitems.NewCreateSceneItemParams().
			WithSceneName(sceneName).
			WithSourceName(sourceName).
			WithSceneItemEnabled(true))
		if err != nil {
			b.lastErr = err.Error()
			return nil, err
		}
		_, _ = client.Inputs.SetInputSettings(inputs.NewSetInputSettingsParams().
			WithInputName(sourceName).
			WithInputSettings(browserSettings(overlayURL)))
		return &EnsureResult{
			Action:     "added_existing",
			Scene:      sceneName,
			SourceName: sourceName,
			Message:    "Источник уже был в OBS — добавлен на выбранную сцену без дубля",
		}, nil
	}

	enabled := true
	_, err = client.Inputs.CreateInput(inputs.NewCreateInputParams().
		WithSceneName(sceneName).
		WithInputName(sourceName).
		WithInputKind("browser_source").
		WithSceneItemEnabled(enabled).
		WithInputSettings(browserSettings(overlayURL)))
	if err != nil {
		b.lastErr = err.Error()
		return nil, err
	}
	// Keep selected scene in memory.
	b.cfg.Scene = sceneName
	return &EnsureResult{
		Action:     "created",
		Scene:      sceneName,
		SourceName: sourceName,
		Message:    "Browser Source создан на выбранной сцене",
	}, nil
}

func browserSettings(overlayURL string) map[string]any {
	return map[string]any{
		"url":                 overlayURL,
		"width":               470,
		"height":              120,
		"css":                 "",
		"shutdown":            false,
		"restart_when_active": true,
		"fps":                 30,
	}
}
