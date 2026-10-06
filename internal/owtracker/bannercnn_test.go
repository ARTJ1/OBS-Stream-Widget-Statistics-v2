package owtracker

import (
	"path/filepath"
	"strings"
	"testing"
)

// The embedded model on real captures from an earlier session (1440p, never
// used for training). win_4_flash is the victory fly-in flash: a known
// single-frame weakness that the vote rule turns into a short delay
// (TestVoteVictoryFlashNeverCountsDefeat), so it is only logged here.
func TestEmbeddedBannerCNNSamples(t *testing.T) {
	m, err := loadEmbeddedBannerCNN()
	if err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob("testdata/banners/*.png")
	// banners_cnn: CNN-only samples (other languages / colours the legacy reader can't do).
	cnnOnly, _ := filepath.Glob("testdata/banners_cnn/*.png")
	for _, f := range append(files, cnnOnly...) {
		name := filepath.Base(f)
		n, w, l := m.Classify(loadSample(t, f))
		got := Outcome("")
		switch {
		case w >= cnnStrongP:
			got = OutcomeWin
		case l >= cnnStrongP:
			got = OutcomeLoss
		}
		want := Outcome("")
		switch {
		case strings.HasPrefix(name, "win"):
			want = OutcomeWin
		case strings.HasPrefix(name, "loss"):
			want = OutcomeLoss
		}
		if strings.Contains(name, "flash") {
			t.Logf("%s (known weakness): none=%.2f win=%.2f loss=%.2f", name, n, w, l)
			continue
		}
		if got != want {
			t.Errorf("%s: got %q want %q (none=%.2f win=%.2f loss=%.2f)", name, got, want, n, w, l)
		}
	}
}

func BenchmarkBannerCNN(b *testing.B) {
	m, err := loadEmbeddedBannerCNN()
	if err != nil {
		b.Fatal(err)
	}
	zone := loadSampleB(b, "testdata/banners/win_1.png")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Classify(zone)
	}
}
