//go:build !windows

package runtimeinfo

import (
	"fmt"
	"os"
	"syscall"
)

func flock(f *os.File) error {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("another instance is already running: %w", err)
	}
	return nil
}
