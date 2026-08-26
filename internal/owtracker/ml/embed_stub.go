//go:build !windows

package ml

func embeddedMLP() ([]byte, error) { return nil, nil }
