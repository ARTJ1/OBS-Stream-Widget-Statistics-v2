package tray

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/getlantern/systray"
)

type UpdateNotice struct {
	Available bool
	Latest    string
}

type Hooks struct {
	AdminURL string
	OnQuit   func()
	Updates  <-chan UpdateNotice
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
	mUpdate := systray.AddMenuItem("No updates", "Open admin update page")
	mUpdate.Hide()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop widget server")

	updateURL := strings.TrimRight(hooks.AdminURL, "/") + "/?update=1"

	go func() {
		for {
			select {
			case u, ok := <-hooks.Updates:
				if !ok {
					continue
				}
				if u.Available && u.Latest != "" {
					mUpdate.SetTitle("Update " + u.Latest)
					mUpdate.Show()
					systray.SetIcon(alertIconBytes())
					systray.SetTooltip("Widget Stats · update " + u.Latest)
				} else {
					mUpdate.Hide()
					systray.SetIcon(iconBytes())
					systray.SetTooltip("OBS Stream Widget Statistics v2")
				}
			case <-mAdmin.ClickedCh:
				_ = openURL(hooks.AdminURL)
			case <-mUpdate.ClickedCh:
				_ = openURL(updateURL)
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
