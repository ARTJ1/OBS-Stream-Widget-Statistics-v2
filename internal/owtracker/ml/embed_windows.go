//go:build windows

package ml

import (
	"embed"
	"io/fs"
)

//go:embed assets/models/banner_mlp.json
var embeddedModel embed.FS

func embeddedMLP() ([]byte, error) {
	return fs.ReadFile(embeddedModel, "assets/models/banner_mlp.json")
}
