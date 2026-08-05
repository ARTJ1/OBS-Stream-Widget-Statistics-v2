package runtimeinfo

import "os"

func lockFile(f *os.File) error {
	return flock(f)
}
