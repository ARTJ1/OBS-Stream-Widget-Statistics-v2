//go:build windows

package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func scheduleWindowsReplace(exe, newExe string, pid int) error {
	bat := exe + ".update.bat"
	// No "timeout" (can flash). ping is silent with redirect.
	content := fmt.Sprintf(`@echo off
setlocal
set "EXE=%s"
set "NEW=%s"
set "PID=%d"
:wait
tasklist /FI "PID eq %%PID%%" 2>NUL | find "%%PID%%" >NUL
if not errorlevel 1 (
  ping -n 2 127.0.0.1 >NUL
  goto wait
)
move /Y "%%NEW%%" "%%EXE%%" >NUL
start "" "%%EXE%%"
del "%%~f0"
`, exe, newExe, pid)
	if err := os.WriteFile(bat, []byte(content), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("cmd.exe", "/C", bat)
	cmd.Dir = filepath.Dir(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	return cmd.Start()
}
