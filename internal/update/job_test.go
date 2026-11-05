package update

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDownloadReportsProgress(t *testing.T) {
	body := strings.Repeat("x", 64*1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "65536")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	var lastDone, lastTotal int64
	dest := filepath.Join(t.TempDir(), "f.exe")
	err := downloadOnce(srv.Client(), srv.URL, dest, func(done, total int64) {
		lastDone, lastTotal = done, total
	})
	if err != nil {
		t.Fatal(err)
	}
	if lastDone != 65536 || lastTotal != 65536 {
		t.Fatalf("progress %d of %d", lastDone, lastTotal)
	}
}

func TestDownloadStallIsReported(t *testing.T) {
	old := stallTimeout
	stallTimeout = 200 * time.Millisecond
	defer func() { stallTimeout = old }()

	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "65536")
		_, _ = w.Write(make([]byte, 2048))
		w.(http.Flusher).Flush()
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	err := downloadOnce(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "f.exe"), nil)
	if failureCode(err) != CodeStalled {
		t.Fatalf("got %v", err)
	}
}

func TestJobFailsWhenNothingToInstall(t *testing.T) {
	j := NewJob(t.TempDir())
	j.Start(func() Info { return Info{Error: "offline"} }, t.TempDir(), func() { t.Error("must not quit") })
	deadline := time.Now().Add(5 * time.Second)
	for j.Status().Phase != PhaseFailed && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if p := j.Status(); p.Phase != PhaseFailed || p.Code != CodeDownload || p.Detail != "offline" {
		t.Fatalf("got %+v", p)
	}
	j.Dismiss()
	if p := j.Status(); p.Phase != PhaseIdle {
		t.Fatalf("after dismiss: %+v", p)
	}
}

func TestJobPicksUpFailedInstall(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, swapResultFile), []byte(CodeLocked), 0o644); err != nil {
		t.Fatal(err)
	}
	if p := NewJob(dir).Status(); p.Phase != PhaseFailed || p.Code != CodeLocked {
		t.Fatalf("got %+v", p)
	}
	if p := NewJob(dir).Status(); p.Phase != PhaseIdle {
		t.Fatalf("second run: %+v", p)
	}
}
