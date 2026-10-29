package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReleaseNotes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/"+GitHubOwner+"/"+GitHubRepo+"/releases/tags/v3.0.0" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"name":" v3.0.0 — big ","body":"line 1\r\n- item"}`))
	}))
	defer srv.Close()
	old := apiBase
	apiBase = srv.URL
	defer func() { apiBase = old }()

	c := NewChecker("v2.8.1", "")
	name, body, err := c.releaseNotes("v3.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if name != "v3.0.0 — big" || body != "line 1\n- item" {
		t.Fatalf("got %q / %q", name, body)
	}
	if _, _, err := c.releaseNotes("v9.9.9"); err == nil {
		t.Fatal("want error for a missing release")
	}
}
