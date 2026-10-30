package server

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// maxFontBytes caps one uploaded font file; real-world TTF/OTF families rarely exceed a few MB.
const maxFontBytes = 15 << 20

var fontExts = map[string]bool{".ttf": true, ".otf": true, ".woff": true, ".woff2": true}

func init() {
	// Windows registry often lacks these, and some browsers refuse fonts served as text/plain.
	_ = mime.AddExtensionType(".woff2", "font/woff2")
	_ = mime.AddExtensionType(".woff", "font/woff")
	_ = mime.AddExtensionType(".ttf", "font/ttf")
	_ = mime.AddExtensionType(".otf", "font/otf")
}

type customFont struct {
	Family string `json:"family"`
	File   string `json:"file"`
	URL    string `json:"url"`
}

func (s *Server) fontsDir() string {
	return filepath.Join(s.Store.Dir(), "uploads", "fonts")
}

// fontFamilyName turns a file name into a CSS-safe family: letters, digits, spaces, '-' and '_'.
func fontFamilyName(filename string) string {
	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	var b strings.Builder
	for _, r := range base {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ', r == '.':
			b.WriteRune(' ')
		}
	}
	name := strings.Join(strings.Fields(b.String()), " ")
	if len([]rune(name)) > 60 {
		name = string([]rune(name)[:60])
	}
	return name
}

// fontMagicOK checks the file header so a renamed image or script is not served as a font.
func fontMagicOK(head []byte) bool {
	if len(head) < 4 {
		return false
	}
	sig := head[:4]
	return bytes.Equal(sig, []byte{0, 1, 0, 0}) ||
		bytes.Equal(sig, []byte("true")) ||
		bytes.Equal(sig, []byte("OTTO")) ||
		bytes.Equal(sig, []byte("wOFF")) ||
		bytes.Equal(sig, []byte("wOF2"))
}

func (s *Server) listFonts() []customFont {
	entries, err := os.ReadDir(s.fontsDir())
	if err != nil {
		return []customFont{}
	}
	out := make([]customFont, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !fontExts[strings.ToLower(filepath.Ext(e.Name()))] {
			continue
		}
		family := fontFamilyName(e.Name())
		if family == "" {
			continue
		}
		out = append(out, customFont{Family: family, File: e.Name(), URL: "/uploads/fonts/" + e.Name()})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Family) < strings.ToLower(out[j].Family) })
	return out
}

func (s *Server) fontsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"fonts": s.listFonts()})
	case http.MethodPost:
		s.uploadFont(w, r)
	case http.MethodDelete:
		s.deleteFont(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) uploadFont(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFontBytes+(1<<20))
	if err := r.ParseMultipartForm(maxFontBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "font file too large or broken"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file required"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	family := fontFamilyName(header.Filename)
	if !fontExts[ext] || family == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported font type (TTF / OTF / WOFF / WOFF2)"})
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if !fontMagicOK(data) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file is not a font"})
		return
	}

	dir := s.fontsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// One file per family: re-uploading "Furore" in another format replaces the old one.
	for _, f := range s.listFonts() {
		if strings.EqualFold(f.Family, family) {
			_ = os.Remove(filepath.Join(dir, f.File))
		}
	}
	name := family + ext
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"font":  customFont{Family: family, File: name, URL: "/uploads/fonts/" + name},
		"fonts": s.listFonts(),
	})
}

func (s *Server) deleteFont(w http.ResponseWriter, r *http.Request) {
	file := filepath.Base(r.URL.Query().Get("file"))
	if file == "" || file == "." || !fontExts[strings.ToLower(filepath.Ext(file))] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file required"})
		return
	}
	if err := os.Remove(filepath.Join(s.fontsDir(), file)); err != nil && !os.IsNotExist(err) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "fonts": s.listFonts()})
}
