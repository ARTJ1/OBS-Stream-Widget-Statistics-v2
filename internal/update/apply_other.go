//go:build !windows

package update

import "fmt"

func scheduleWindowsReplace(exe, newExe string, pid int) error {
	return fmt.Errorf("auto-apply is only supported on Windows")
}
