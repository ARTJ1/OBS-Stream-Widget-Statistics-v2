package server_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/hub"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/obsbridge"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/runtimeinfo"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/server"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/skins"
	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
	webassets "github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/web"
)

type fontsResp struct {
	Fonts []struct {
		Family string `json:"family"`
		File   string `json:"file"`
		URL    string `json:"url"`
	} `json:"fonts"`
}

func uploadFontReq(t *testing.T, h http.Handler, filename string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(data)
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/fonts", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCustomFontsUploadListDelete(t *testing.T) {
	dir := t.TempDir()
	st, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	skinStore, err := skins.New(dir)
	if err != nil {
		t.Fatal(err)
	}
	info := runtimeinfo.Build("127.0.0.1", 19124, "test")
	h := server.New(st, skinStore, hub.New(), obsbridge.New(), info, webassets.FS, nil).Handler()

	otf := append([]byte("OTTO"), make([]byte, 64)...)
	if rec := uploadFontReq(t, h, "Furore.otf", otf); rec.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body.String())
	}
	// A renamed image must be rejected even with a font extension.
	if rec := uploadFontReq(t, h, "evil.ttf", []byte("\x89PNG\r\n\x1a\n")); rec.Code != http.StatusBadRequest {
		t.Fatalf("png as ttf: want 400, got %d", rec.Code)
	}
	if rec := uploadFontReq(t, h, "script.js", otf); rec.Code != http.StatusBadRequest {
		t.Fatalf("js ext: want 400, got %d", rec.Code)
	}
	// Re-upload of the same family in another format replaces the old file.
	woff2 := append([]byte("wOF2"), make([]byte, 64)...)
	if rec := uploadFontReq(t, h, "Furore.woff2", woff2); rec.Code != http.StatusOK {
		t.Fatalf("re-upload: %d", rec.Code)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/fonts", nil))
	var list fontsResp
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Fonts) != 1 || list.Fonts[0].Family != "Furore" || list.Fonts[0].File != "Furore.woff2" {
		t.Fatalf("list: %+v", list.Fonts)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, list.Fonts[0].URL, nil))
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "font/woff2" {
		t.Fatalf("serve: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}

	// Path traversal in delete is reduced to a base name inside the fonts folder.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/fonts?file=..%2F..%2Fsettings.json", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("traversal delete: want 400, got %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/fonts?file=Furore.woff2", nil))
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || len(list.Fonts) != 0 {
		t.Fatalf("delete: %d %+v", rec.Code, list.Fonts)
	}
}
