package tray

import _ "embed"

//go:embed icons/app.ico
var appIcon []byte

func iconBytes() []byte {
	return appIcon
}
