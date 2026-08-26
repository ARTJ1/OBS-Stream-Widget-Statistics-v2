//go:build !windows

package ml

func embeddedTextMLP() ([]byte, error) {
	return nil, nil
}
