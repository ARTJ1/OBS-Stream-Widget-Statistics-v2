//go:build windows

package ml

import (
	"embed"
	"io/fs"
)

//go:embed assets/models/banner_text_mlp.json
var embeddedTextModel embed.FS

func embeddedTextMLP() ([]byte, error) {
	return fs.ReadFile(embeddedTextModel, "assets/models/banner_text_mlp.json")
}
