//go:build windows

package owtracker

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func overwatchProcessRunning() bool {
	return len(overwatchPIDs()) > 0
}

func overwatchPIDs() map[uint32]struct{} {
	out := map[uint32]struct{}{}
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return out
	}
	defer windows.CloseHandle(snap)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	for err := windows.Process32First(snap, &entry); err == nil; err = windows.Process32Next(snap, &entry) {
		name := strings.ToLower(windows.UTF16ToString(entry.ExeFile[:]))
		if name == "overwatch.exe" {
			out[entry.ProcessID] = struct{}{}
		}
	}
	return out
}

func overwatchGameBounds() (x, y, w, h int, ok bool) {
	hwnd := findOverwatchHWND()
	if hwnd == 0 {
		return 0, 0, 0, 0, false
	}
	return clientBoundsOf(hwnd)
}

func clientBoundsOf(hwnd uintptr) (x, y, w, h int, ok bool) {
	var rc struct{ Left, Top, Right, Bottom int32 }
	r, _, _ := procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	if r == 0 {
		return 0, 0, 0, 0, false
	}
	pt := struct{ X, Y int32 }{X: rc.Left, Y: rc.Top}
	_, _, _ = procClientToScreen.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
	w = int(rc.Right - rc.Left)
	h = int(rc.Bottom - rc.Top)
	if w <= 32 || h <= 32 {
		return 0, 0, 0, 0, false
	}
	return int(pt.X), int(pt.Y), w, h, true
}

func findOverwatchHWND() uintptr {
	pids := overwatchPIDs()
	if len(pids) == 0 {
		return 0
	}
	type cand struct {
		hwnd  uintptr
		area  int
		named bool
	}
	best := cand{}
	cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
		vis, _, _ := procIsWindowVisible.Call(hwnd)
		if vis == 0 {
			return 1
		}
		var pid uint32
		_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if _, ok := pids[pid]; !ok {
			return 1
		}
		x, y, w, h, ok := clientBoundsOf(hwnd)
		if !ok {
			return 1
		}
		_ = x
		_ = y
		named := isOverwatchTitle(windowTitle(hwnd))
		area := w * h
		if named && (!best.named || area > best.area) {
			best = cand{hwnd: hwnd, area: area, named: true}
		} else if !best.named && area > best.area {
			best = cand{hwnd: hwnd, area: area, named: false}
		}
		return 1
	})
	_, _, _ = procEnumWindows.Call(cb, 0)
	return best.hwnd
}
