//go:build windows

package owtracker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func verifyTesseract(exe string) bool {
	cmd := exec.Command(exe, "--version")
	configureHiddenCmd(cmd)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	return cmd.Run() == nil
}

func runTesseract(imagePath string) (string, error) {
	return runTesseractPSM(imagePath, "7")
}

func runTesseractPSM(imagePath, psm string) (string, error) {
	if ocrExePath == "" {
		return "", fmt.Errorf("embedded tesseract missing")
	}
	absImg, err := filepath.Abs(imagePath)
	if err != nil {
		return "", err
	}
	text, err := execTesseract(absImg, psm, true)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(text) != "" {
		return text, nil
	}
	return execTesseract(absImg, psm, false)
}

func execTesseract(absImg, psm string, whitelist bool) (string, error) {
	args := []string{absImg, "stdout", "-l", "rus+eng", "--psm", psm, "--oem", "1"}
	if whitelist {
		args = append(args, "-c", "tessedit_char_whitelist=АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ!ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	cmd := exec.Command(ocrExePath, args...)
	configureHiddenCmd(cmd)
	cmd.Dir = filepath.Dir(ocrExePath)
	if td := tessdataDir(ocrExePath); td != "" {
		cmd.Env = append(os.Environ(), "TESSDATA_PREFIX="+td)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return strings.TrimSpace(out.String()), nil
}

func tessdataDir(exePath string) string {
	p := filepath.Join(filepath.Dir(exePath), "tessdata")
	if st, err := os.Stat(p); err == nil && st.IsDir() {
		if _, err := os.Stat(filepath.Join(p, "eng.traineddata")); err == nil {
			return p + string(os.PathSeparator)
		}
	}
	return ""
}
