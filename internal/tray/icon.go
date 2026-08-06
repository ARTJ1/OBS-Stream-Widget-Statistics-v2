package tray

import _ "embed"

//go:embed icons/app.ico
var appIcon []byte

//go:embed icons/app_alert.ico
var appAlertIcon []byte

func iconBytes() []byte {
	return appIcon
}

func alertIconBytes() []byte {
	return appAlertIcon
}
