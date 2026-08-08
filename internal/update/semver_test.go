package update

import "testing"

func TestNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v2.2.0", "v2.1.2", true},
		{"v2.1.2", "v2.1.2", false},
		{"v2.1.1", "v2.1.2", false},
		{"streamdock-icons-v1.0.0", "v2.1.2", false},
		{"v2.2.0", "dev", false},
		{"", "v2.1.2", false},
	}
	for _, tc := range cases {
		if got := Newer(tc.latest, tc.current); got != tc.want {
			t.Fatalf("Newer(%q, %q)=%v want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestTagFromReleaseURL(t *testing.T) {
	got := tagFromReleaseURL("https://github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/releases/tag/v2.2.0")
	if got != "v2.2.0" {
		t.Fatalf("got %q", got)
	}
}
