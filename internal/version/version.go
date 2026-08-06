package version

// Version is set at build time via:
//
//	-ldflags "-X github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/version.Version=v2.0.6"
//
// Default "dev" skips update prompts.
var Version = "dev"

func Display() string {
	if Version == "" {
		return "dev"
	}
	return Version
}
