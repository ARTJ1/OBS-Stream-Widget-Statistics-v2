package owtracker

import (
	"bufio"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Windows OCR (Windows.Media.Ocr) through one long-lived hidden PowerShell process.
// Built into Windows 10/11: nothing to download, no training.

//go:embed assets/winocr.ps1
var winOCRScript []byte

const winOCRTimeout = 8 * time.Second

var errWinOCRUnavailable = errors.New("windows ocr unavailable")

type winOCR struct {
	dataDir string

	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	lines  chan string
	langs  []string
	nextID int
	err    string
}

func newWinOCR(dataDir string) *winOCR {
	return &winOCR{dataDir: dataDir}
}

// Langs reports OCR languages usable for banner words (after the first start).
func (o *winOCR) Langs() []string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]string(nil), o.langs...)
}

func (o *winOCR) LastError() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.err
}

func (o *winOCR) Running() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.cmd != nil
}

func (o *winOCR) scriptPath() (string, error) {
	dir := filepath.Join(o.dataDir, ".ocr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	p := filepath.Join(dir, "winocr.ps1")
	// UTF-8 BOM so Windows PowerShell 5.1 never misreads the script.
	body := append([]byte{0xEF, 0xBB, 0xBF}, winOCRScript...)
	if cur, err := os.ReadFile(p); err != nil || string(cur) != string(body) {
		if err := os.WriteFile(p, body, 0o644); err != nil {
			return "", err
		}
	}
	return p, nil
}

// start: caller holds o.mu.
func (o *winOCR) start() error {
	if o.cmd != nil {
		return nil
	}
	script, err := o.scriptPath()
	if err != nil {
		return err
	}
	cmd := exec.Command("powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive",
		"-ExecutionPolicy", "Bypass", "-File", script)
	configureHiddenCmd(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	lines := make(chan string, 64)
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 64<<10), 1<<20)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
		_ = cmd.Wait()
	}()

	select {
	case line, ok := <-lines:
		if !ok || !strings.HasPrefix(line, "LANGS") {
			_ = cmd.Process.Kill()
			return fmt.Errorf("%w: bad handshake %q", errWinOCRUnavailable, line)
		}
		o.langs = nil
		for _, l := range strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "LANGS")), ",") {
			if l = strings.TrimSpace(l); l != "" {
				o.langs = append(o.langs, l)
			}
		}
	case <-time.After(20 * time.Second):
		_ = cmd.Process.Kill()
		return fmt.Errorf("%w: startup timeout", errWinOCRUnavailable)
	}
	o.cmd, o.stdin, o.lines = cmd, stdin, lines
	log.Printf("owtracker: windows ocr ready (langs=%s)", strings.Join(o.langs, ","))
	return nil
}

// stop: caller holds o.mu.
func (o *winOCR) stop() {
	if o.cmd == nil {
		return
	}
	_ = o.stdin.Close()
	_ = o.cmd.Process.Kill()
	o.cmd, o.stdin, o.lines = nil, nil, nil
}

func (o *winOCR) Close() {
	o.mu.Lock()
	o.stop()
	o.mu.Unlock()
}

// Recognize returns OCR text per language for a PNG file.
func (o *winOCR) Recognize(pngPath string) (map[string]string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.start(); err != nil {
		o.err = err.Error()
		return nil, err
	}
	abs, err := filepath.Abs(pngPath)
	if err != nil {
		return nil, err
	}
	o.nextID++
	id := strconv.Itoa(o.nextID)
	if _, err := fmt.Fprintf(o.stdin, "REQ %s %s\n", id, abs); err != nil {
		o.stop()
		o.err = err.Error()
		return nil, err
	}
	out := map[string]string{}
	deadline := time.After(winOCRTimeout)
	for {
		select {
		case line, ok := <-o.lines:
			if !ok {
				o.stop()
				o.err = "ocr process exited"
				return nil, errWinOCRUnavailable
			}
			f := strings.SplitN(line, " ", 4)
			if len(f) < 2 || f[1] != id {
				continue // stale line from a timed-out request
			}
			switch f[0] {
			case "RES":
				if len(f) == 4 {
					out[f[2]] = decodeB64(f[3])
				} else if len(f) == 3 {
					out[f[2]] = ""
				}
			case "ERR":
				msg := ""
				if len(f) >= 3 {
					msg = decodeB64(strings.Join(f[2:], " "))
				}
				o.err = msg
			case "END":
				if len(out) > 0 {
					o.err = ""
				}
				return out, nil
			}
		case <-deadline:
			o.stop()
			o.err = "ocr timeout"
			return nil, fmt.Errorf("%w: timeout", errWinOCRUnavailable)
		}
	}
}

func decodeB64(s string) string {
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return ""
	}
	return string(b)
}
