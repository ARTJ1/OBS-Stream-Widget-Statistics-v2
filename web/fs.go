package web

import "embed"

//go:embed all:overlay all:admin
var FS embed.FS
