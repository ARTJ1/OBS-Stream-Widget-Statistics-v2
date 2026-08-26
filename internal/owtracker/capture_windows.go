//go:build windows

// Screen capture and foreground-window title (robotgo.GetTitle / CaptureImg / GetScreenSize
// equivalents). robotgo is CGO; this project ships a CGO-free windowsgui exe, so we use
// Win32 + kbinani/screenshot instead.

package owtracker

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/kbinani/screenshot"
	"golang.org/x/sys/windows"
)

var (
	user32 = windows.NewLazySystemDLL("user32.dll")
	shcore = windows.NewLazySystemDLL("shcore.dll")

	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetClientRect            = user32.NewProc("GetClientRect")
	procClientToScreen           = user32.NewProc("ClientToScreen")
	procSetProcessDpiAwareness   = shcore.NewProc("SetProcessDpiAwareness")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

func initDPI() {
	// PROCESS_PER_MONITOR_DPI_AWARE = 2 so % boxes match physical pixels.
	_, _, _ = procSetProcessDpiAwareness.Call(2)
}

func activeWindowTitle() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}
	n, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	_, _, _ = procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(n+1))
	return windows.UTF16ToString(buf)
}

func screenSize() (int, int) {
	b := screenshot.GetDisplayBounds(0)
	return b.Dx(), b.Dy()
}

func isOverwatchForeground() bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}
	return hwndIsOverwatchGame(hwnd)
}

func hwndIsOverwatchGame(hwnd uintptr) bool {
	if isOverwatchTitle(windowTitle(hwnd)) {
		return true
	}
	// Exclusive fullscreen often has an empty caption; match by process instead.
	var pid uint32
	_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	_, ok := overwatchPIDs()[pid]
	return ok
}

func overwatchClientBounds() (x, y, w, h int, ok bool) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return 0, 0, 0, 0, false
	}
	if !hwndIsOverwatchGame(hwnd) {
		return 0, 0, 0, 0, false
	}
	return clientBoundsOf(hwnd)
}

func foregroundClientBounds() (x, y, w, h int, ok bool) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return 0, 0, 0, 0, false
	}
	return clientBoundsOf(hwnd)
}

func foregroundClientBoundsWithTitle() (x, y, w, h int, title string, ok bool) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return 0, 0, 0, 0, "", false
	}
	x, y, w, h, ok = clientBoundsOf(hwnd)
	if !ok {
		return 0, 0, 0, 0, "", false
	}
	return x, y, w, h, windowTitle(hwnd), true
}

func windowTitle(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	_, _, _ = procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(n+1))
	return windows.UTF16ToString(buf)
}

func captureRegion(x, y, w, h int) (image.Image, error) {
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("owtracker: invalid capture size %dx%d", w, h)
	}
	img, err := screenshot.Capture(x, y, w, h)
	if err != nil {
		return nil, err
	}
	return img, nil
}
