package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/hub"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/obsbridge"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/runtimeinfo"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/skins"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
	"github.com/gorilla/websocket"
)

type Server struct {
	Store   *store.Store
	Skins   *skins.Store
	Hub     *hub.Hub
	OBS     *obsbridge.Bridge
	Runtime runtimeinfo.Info
	WebFS   fs.FS
	mux     *http.ServeMux
}

func New(st *store.Store, skinStore *skins.Store, h *hub.Hub, obs *obsbridge.Bridge, info runtimeinfo.Info, webFS fs.FS) *Server {
	s := &Server{Store: st, Skins: skinStore, Hub: h, OBS: obs, Runtime: info, WebFS: webFS, mux: http.NewServeMux()}
	s.syncOBSFromSettings()
	s.routes()
	return s
}

func (s *Server) syncOBSFromSettings() {
	cfg := s.Store.Settings()
	s.OBS.Configure(obsbridge.Config{
		Host:       cfg.ObsHost,
		Port:       cfg.ObsPort,
		Password:   cfg.ObsPassword,
		Scene:      cfg.ObsScene,
		SourceName: cfg.ObsSourceName,
	})
}

func (s *Server) Handler() http.Handler {
	return s.cors(s.mux)
}

func (s *Server) Serve(ln net.Listener) error {
	return http.Serve(ln, s.Handler())
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/win", s.methodAction(s.Store.AddWin, "win"))
	s.mux.HandleFunc("/api/loss", s.methodAction(s.Store.AddLoss, "loss"))
	s.mux.HandleFunc("/api/rank/up", s.methodAction(s.Store.RankUp, "rank"))
	s.mux.HandleFunc("/api/rank/down", s.methodAction(s.Store.RankDown, "rank"))
	s.mux.HandleFunc("/api/rank/set", s.setRank)
	s.mux.HandleFunc("/api/reset", s.methodAction(s.Store.Reset, "reset"))
	s.mux.HandleFunc("/api/state", s.stateHandler)
	s.mux.HandleFunc("/api/settings", s.settings)
	s.mux.HandleFunc("/api/runtime", s.getRuntime)
	s.mux.HandleFunc("/api/snapshot", s.getSnapshot)
	s.mux.HandleFunc("/api/obs/status", s.obsStatus)
	s.mux.HandleFunc("/api/obs/connect", s.obsConnect)
	s.mux.HandleFunc("/api/obs/disconnect", s.obsDisconnect)
	s.mux.HandleFunc("/api/obs/scenes", s.obsScenes)
	s.mux.HandleFunc("/api/obs/ensure", s.obsEnsure)
	s.mux.HandleFunc("/api/upload/bg", s.uploadBg)
	s.mux.HandleFunc("/api/skins", s.customSkins)
	s.mux.HandleFunc("/api/show", s.showOverlay)
	s.mux.HandleFunc("/ws", s.ws)
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	overlayFS, err := fs.Sub(s.WebFS, "overlay")
	if err == nil {
		s.mux.Handle("/overlay/", http.StripPrefix("/overlay/", http.FileServer(http.FS(overlayFS))))
	}
	adminFS, err := fs.Sub(s.WebFS, "admin")
	if err == nil {
		s.mux.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.FS(adminFS))))
	}
	uploadsDir := filepath.Join(s.Store.Dir(), "uploads")
	_ = os.MkdirAll(uploadsDir, 0o755)
	s.mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/admin/", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
}

type actionFn func() (store.Snapshot, error)

func (s *Server) methodAction(fn actionFn, msgType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		snap, err := fn()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Hub.Broadcast(msgType, snap)
		writeJSON(w, http.StatusOK, snap)
	}
}

// ApplyHotkey runs the same mutations as /api/* for the OBS file-drop hotkey path
// (no HTTP / no console from Lua).
func (s *Server) ApplyHotkey(name string) {
	var fn actionFn
	var msgType string
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "win":
		fn, msgType = s.Store.AddWin, "win"
	case "loss":
		fn, msgType = s.Store.AddLoss, "loss"
	case "rank_up":
		fn, msgType = s.Store.RankUp, "rank"
	case "rank_down":
		fn, msgType = s.Store.RankDown, "rank"
	case "reset":
		fn, msgType = s.Store.Reset, "reset"
	default:
		return
	}
	snap, err := fn()
	if err != nil {
		log.Printf("hotkey %s: %v", name, err)
		return
	}
	s.Hub.Broadcast(msgType, snap)
}

func (s *Server) getState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.Store.State())
}

func (s *Server) stateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getState(w, r)
	case http.MethodPut, http.MethodPost:
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var st store.State
		if err := json.Unmarshal(body, &st); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		snap, err := s.Store.SetState(st)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Hub.Broadcast("state", snap)
		writeJSON(w, http.StatusOK, snap)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) setRank(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rank := -1
	if q := r.URL.Query().Get("rank"); q != "" {
		var n int
		if _, err := fmt.Sscanf(q, "%d", &n); err == nil {
			rank = n
		}
	}
	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if len(body) > 0 {
			var req struct {
				Rank int `json:"rank"`
			}
			if json.Unmarshal(body, &req) == nil {
				rank = req.Rank
			}
		}
	}
	if rank < 0 {
		http.Error(w, "rank required", http.StatusBadRequest)
		return
	}
	snap, err := s.Store.SetRank(rank)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Hub.Broadcast("rank", snap)
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) getSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.Store.Snapshot())
}

func (s *Server) getRuntime(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.Runtime)
}

func (s *Server) settings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.Store.Settings())
	case http.MethodPut, http.MethodPost:
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var next store.Settings
		if err := json.Unmarshal(body, &next); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		snap, err := s.Store.UpdateSettings(next)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.syncOBSFromSettings()
		s.Hub.Broadcast("settings", snap)
		writeJSON(w, http.StatusOK, snap)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) uploadBg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
	default:
		http.Error(w, "unsupported image type", http.StatusBadRequest)
		return
	}

	uploadsDir := filepath.Join(s.Store.Dir(), "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name := fmt.Sprintf("bg_%d%s", time.Now().UnixNano(), ext)
	dstPath := filepath.Join(uploadsDir, name)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := "/uploads/" + name
	cfg := s.Store.Settings()
	cfg.BgImage = url
	cfg.BgType = "image"
	snap, err := s.Store.UpdateSettings(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.Hub.Broadcast("settings", snap)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"url":      url,
		"settings": snap.Settings,
		"state":    snap.State,
	})
}

func (s *Server) obsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.OBS.Status())
}

func (s *Server) obsConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.syncOBSFromSettings()
	if err := s.OBS.Connect(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":      false,
			"error":   err.Error(),
			"status":  s.OBS.Status(),
			"hint":    "Включи Tools → WebSocket Server в OBS (порт 4455) и укажи пароль в админке.",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": s.OBS.Status()})
}

func (s *Server) obsDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.OBS.Disconnect()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": s.OBS.Status()})
}

func (s *Server) obsScenes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.syncOBSFromSettings()
	names, current, err := s.OBS.Scenes()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"scenes":         names,
		"currentProgram": current,
		"selected":       s.Store.Settings().ObsScene,
	})
}

func (s *Server) obsEnsure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.syncOBSFromSettings()
	scene := s.Store.Settings().ObsScene
	if r.URL.Query().Get("scene") != "" {
		scene = r.URL.Query().Get("scene")
	}
	if r.Method == http.MethodPost {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if len(body) > 0 {
			var req struct {
				Scene string `json:"scene"`
			}
			if json.Unmarshal(body, &req) == nil && req.Scene != "" {
				scene = req.Scene
			}
		}
	}
	res, err := s.OBS.EnsureOnScene(scene, s.Runtime.OverlayURL)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	// Persist selected scene if ensure succeeded.
	cfg := s.Store.Settings()
	if cfg.ObsScene != scene {
		cfg.ObsScene = scene
		_, _ = s.Store.UpdateSettings(cfg)
		s.syncOBSFromSettings()
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": res})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	snap := s.Store.Snapshot()
	_ = conn.WriteJSON(hub.Message{Type: "hello", State: snap.State, Settings: snap.Settings})

	ch := s.Hub.Subscribe()
	defer s.Hub.Unsubscribe(ch)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) showOverlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snap := s.Store.Snapshot()
	s.Hub.BroadcastShow(snap)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "snapshot": snap})
}

func appearanceOnly(in store.Settings) store.Settings {
	out := in
	out.ObsHost = ""
	out.ObsPort = 0
	out.ObsPassword = ""
	out.ObsScene = ""
	out.ObsSourceName = ""
	return out
}

func (s *Server) customSkins(w http.ResponseWriter, r *http.Request) {
	if s.Skins == nil {
		http.Error(w, "skins unavailable", http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"skins": s.Skins.List()})
	case http.MethodPost, http.MethodPut:
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var req struct {
			Name     string          `json:"name"`
			Settings *store.Settings `json:"settings"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		settings := s.Store.Settings()
		if req.Settings != nil {
			settings = *req.Settings
		}
		skin, err := s.Skins.Save(name, appearanceOnly(settings))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, skin)
	case http.MethodDelete:
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if id == "" {
			var req struct {
				ID string `json:"id"`
			}
			body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
			_ = json.Unmarshal(body, &req)
			id = strings.TrimSpace(req.ID)
		}
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}
		if !s.Skins.Delete(id) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/overlay/") || strings.HasPrefix(r.URL.Path, "/admin/") {
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}
