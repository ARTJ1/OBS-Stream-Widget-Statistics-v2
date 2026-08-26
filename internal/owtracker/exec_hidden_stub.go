//go:build !windows

package owtracker

import "os/exec"

func configureHiddenCmd(cmd *exec.Cmd) {}
