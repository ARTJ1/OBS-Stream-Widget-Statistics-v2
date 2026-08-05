package tray

import (
	"os/exec"
	"runtime"

	"github.com/getlantern/systray"
)

type Hooks struct {
	AdminURL string
	OnQuit   func()
}

func Run(hooks Hooks) {
	systray.Run(func() { onReady(hooks) }, func() {
		if hooks.OnQuit != nil {
			hooks.OnQuit()
		}
	})
}

func onReady(hooks Hooks) {
	systray.SetTitle("Widget Stats")
	systray.SetTooltip("OBS Stream Widget Statistics v2")
	systray.SetIcon(iconBytes())

	mAdmin := systray.AddMenuItem("Open Admin", "Open settings in browser")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop widget server")

	go func() {
		for {
			select {
			case <-mAdmin.ClickedCh:
				_ = openURL(hooks.AdminURL)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func openURL(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
