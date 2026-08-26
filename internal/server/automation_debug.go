package server

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
)

func (s *Server) devAutomationOnly(w http.ResponseWriter) bool {
	if owtracker.DevMode() {
		return true
	}
	http.Error(w, "not found", http.StatusNotFound)
	return false
}

func (s *Server) automationDebug(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.Auto.DebugStatus())
	case http.MethodPost, http.MethodPut:
		q := r.URL.Query()
		if v := strings.TrimSpace(q.Get("testMode")); v != "" {
			s.Auto.SetTestMode(v == "1" || strings.EqualFold(v, "true") || v == "on")
		}
		if v := strings.TrimSpace(q.Get("debugLog")); v != "" {
			s.Auto.SetDebugLog(v == "1" || strings.EqualFold(v, "true") || v == "on")
		}
		if v := strings.TrimSpace(q.Get("dryRun")); v != "" {
			s.Auto.SetDryRun(v == "1" || strings.EqualFold(v, "true") || v == "on")
		}
		if v := strings.TrimSpace(q.Get("captureSource")); v != "" {
			s.Auto.SetCaptureSource(v)
		}
		if v := strings.TrimSpace(q.Get("matchThreshold")); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				s.Auto.SetMatchThreshold(n)
			}
		}
		writeJSON(w, http.StatusOK, s.Auto.DebugStatus())
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) automationDebugProbe(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	probe, err := s.Auto.ProbeLive()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if s.Auto.ApplyIfReady(r.Context(), probe) {
		if last := s.Auto.DebugStatus().LastApplied; last != nil {
			probe = *last
		}
	}
	s.Auto.RecordProbePublic(probe)
	writeJSON(w, http.StatusOK, map[string]any{"probe": probe, "debug": s.Auto.DebugStatus()})
}

func (s *Server) automationDebugTest(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file required"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", "":
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported image type — use PNG or JPG"})
		return
	}

	img, err := owtracker.DecodeImagePublic(io.LimitReader(file, 20<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	probe := s.Auto.ProbeImage(img, "upload")
	s.Auto.RecordProbePublic(probe)
	writeJSON(w, http.StatusOK, map[string]any{"probe": probe, "debug": s.Auto.DebugStatus()})
}

func (s *Server) automationDebugLog(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		writeJSON(w, http.StatusOK, map[string]any{"entries": []any{}, "debug": owtracker.DebugStatus{}})
		return
	}
	limit := 50
	if q := strings.TrimSpace(r.URL.Query().Get("limit")); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 {
			limit = n
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"entries": s.Auto.RecentProbes(limit),
		"debug":   s.Auto.DebugStatus(),
	})
}

func (s *Server) automationDebugClear(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := s.Auto.ClearDebugLog(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "debug": s.Auto.DebugStatus()})
}

func (s *Server) automationDebugTrain(w http.ResponseWriter, r *http.Request) {
	if !s.devAutomationOnly(w) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	label := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("label")))
	path, err := s.Auto.SaveTrainingSample(label)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "path": path, "label": label})
}
