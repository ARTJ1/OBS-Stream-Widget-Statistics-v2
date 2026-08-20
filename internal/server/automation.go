package server

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
)

func (s *Server) automationToggle(w http.ResponseWriter, r *http.Request) {
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if q := strings.TrimSpace(r.URL.Query().Get("enabled")); q != "" {
			s.Auto.SetEnabled(q == "1" || strings.EqualFold(q, "true") || q == "on")
		}
		writeJSON(w, http.StatusOK, s.Auto.Status())
	case http.MethodPost, http.MethodPut:
		if q := strings.TrimSpace(r.URL.Query().Get("enabled")); q != "" {
			s.Auto.SetEnabled(q == "1" || strings.EqualFold(q, "true") || q == "on")
		} else {
			s.Auto.Toggle()
		}
		writeJSON(w, http.StatusOK, s.Auto.Status())
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) automationStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		writeJSON(w, http.StatusOK, owtracker.Status{State: "inactive"})
		return
	}
	writeJSON(w, http.StatusOK, s.Auto.Status())
}

func (s *Server) automationCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.Auto == nil {
		http.Error(w, "automation unavailable", http.StatusServiceUnavailable)
		return
	}
	kind := owtracker.Outcome(strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind"))))
	if kind != owtracker.OutcomeWin && kind != owtracker.OutcomeLoss {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind=win or kind=loss required"})
		return
	}
	del := r.Method == http.MethodDelete || r.URL.Query().Get("delete") == "1"
	if del {
		if err := s.Auto.ClearTemplate(kind); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": kind, "status": s.Auto.Status()})
		return
	}

	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		s.automationImportUpload(w, r, kind)
		return
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{
		"error": "upload a PNG/JPG screenshot as multipart field \"file\"",
	})
}

func (s *Server) automationImportUpload(w http.ResponseWriter, r *http.Request, kind owtracker.Outcome) {
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

	limited := io.LimitReader(file, 20<<20)
	analysis, err := s.Auto.ImportTemplate(kind, limited)
	if err != nil {
		code := http.StatusBadRequest
		if errors.Is(err, owtracker.ErrTemplateTooSimilar) {
			code = http.StatusConflict
		}
		writeJSON(w, code, map[string]any{
			"error":    err.Error(),
			"analysis": analysis,
			"status":   s.Auto.Status(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"kind":     kind,
		"analysis": analysis,
		"status":   s.Auto.Status(),
	})
}
