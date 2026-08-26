//go:build windows

package ml

// Legacy RGB banner model is not embedded in release builds (text net only).
func embeddedMLP() ([]byte, error) {
	return nil, nil
}
