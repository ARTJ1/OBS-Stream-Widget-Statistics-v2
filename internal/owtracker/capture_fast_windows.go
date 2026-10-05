//go:build windows

package owtracker

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// stripGrabber copies a small screen rectangle with one BitBlt into a reused
// DIB section. Much cheaper than screenshot.Capture (no monitor enumeration,
// no per-call DC/bitmap allocation, no image.RGBA conversion).

var (
	gdi32                  = windows.NewLazySystemDLL("gdi32.dll")
	procGetDC              = user32.NewProc("GetDC")
	procReleaseDC          = user32.NewProc("ReleaseDC")
	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procBitBlt             = gdi32.NewProc("BitBlt")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procDeleteDC           = gdi32.NewProc("DeleteDC")
)

const srcCopy = 0x00CC0020

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type stripGrabber struct {
	mu     sync.Mutex
	w, h   int
	memDC  uintptr
	bitmap uintptr
	bits   unsafe.Pointer
}

func (g *stripGrabber) ensure(w, h int) error {
	if g.memDC != 0 && g.w == w && g.h == h {
		return nil
	}
	g.release()
	dc, _, _ := procCreateCompatibleDC.Call(0)
	if dc == 0 {
		return fmt.Errorf("CreateCompatibleDC failed")
	}
	hdr := bitmapInfoHeader{
		Size:     uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:    int32(w),
		Height:   -int32(h), // top-down rows
		Planes:   1,
		BitCount: 32,
	}
	var bits unsafe.Pointer
	bmp, _, _ := procCreateDIBSection.Call(dc, uintptr(unsafe.Pointer(&hdr)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bmp == 0 || bits == nil {
		procDeleteDC.Call(dc)
		return fmt.Errorf("CreateDIBSection failed")
	}
	procSelectObject.Call(dc, bmp)
	g.memDC, g.bitmap, g.bits, g.w, g.h = dc, bmp, bits, w, h
	return nil
}

func (g *stripGrabber) release() {
	if g.bitmap != 0 {
		procDeleteObject.Call(g.bitmap)
	}
	if g.memDC != 0 {
		procDeleteDC.Call(g.memDC)
	}
	g.memDC, g.bitmap, g.bits = 0, 0, nil
}

// grab copies screen (x,y,w,h) and calls fn with BGRA pixels (valid only inside fn).
func (g *stripGrabber) grab(x, y, w, h int, fn func(bgra []byte)) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if err := g.ensure(w, h); err != nil {
		return err
	}
	screenDC, _, _ := procGetDC.Call(0)
	if screenDC == 0 {
		return fmt.Errorf("GetDC failed")
	}
	ok, _, _ := procBitBlt.Call(g.memDC, 0, 0, uintptr(w), uintptr(h), screenDC, uintptr(x), uintptr(y), srcCopy)
	procReleaseDC.Call(0, screenDC)
	if ok == 0 {
		return fmt.Errorf("BitBlt failed")
	}
	fn(unsafe.Slice((*byte)(g.bits), w*h*4))
	return nil
}

var bannerStrip stripGrabber

// grabStrip is the cheap per-tick probe used before any full capture.
func grabStrip(x, y, w, h int, fn func(bgra []byte)) error {
	return bannerStrip.grab(x, y, w, h, fn)
}
