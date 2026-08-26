//go:build windows

package owtracker

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

func configureHiddenCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
}
