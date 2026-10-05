// owbench runs the Overwatch auto win/loss banner reader over image files, to
// check detection and false positives offline (no game needed).
//
//	go run ./cmd/owbench <folder-or-image>...
//
// Files named win*/loss* are scored as expected results, everything else must
// read as nothing. Full screenshots are cropped to the banner zone automatically.
// Frames from a recording: ffmpeg -i match.mp4 -vf fps=2 frames/none_%05d.png
package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: owbench <folder-or-image>...")
		os.Exit(2)
	}
	files := collect(os.Args[1:])
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no images found")
		os.Exit(2)
	}
	work, err := os.MkdirTemp("", "owbench")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(work)
	ocr := owtracker.NewWindowsOCR(work)
	defer ocr.Close()

	var ok, missed, wrong, falsePos, total int
	for _, f := range files {
		img, err := decode(f)
		if err != nil {
			fmt.Printf("%-40s ERROR %v\n", filepath.Base(f), err)
			continue
		}
		r, err := owtracker.ReadBannerWith(ocr, work, img)
		if err != nil {
			fmt.Printf("%-40s ERROR %v\n", filepath.Base(f), err)
			continue
		}
		total++
		want := expected(f)
		got := string(r.Outcome)
		mark := "ok"
		switch {
		case want == got && want != "":
			ok++
		case want == "" && got == "":
			ok++
		case want == "" && got != "":
			falsePos++
			mark = "FALSE POSITIVE"
		case got == "":
			missed++
			mark = "missed"
		default:
			wrong++
			mark = "WRONG"
		}
		fmt.Printf("%-40s %-6s %-14s likely=%-5v band=%2d%% fill=%2d%% text=%q %s\n",
			filepath.Base(f), orDash(got), mark, r.Likely, r.BandPct, r.FillPct, r.Text, r.Note)
	}
	fmt.Printf("\n%d images: %d ok, %d missed, %d WRONG, %d FALSE POSITIVE\n", total, ok, missed, wrong, falsePos)
	if wrong > 0 || falsePos > 0 {
		os.Exit(1)
	}
}

func expected(path string) string {
	n := strings.ToLower(filepath.Base(path))
	switch {
	case strings.HasPrefix(n, "win"):
		return "win"
	case strings.HasPrefix(n, "loss"):
		return "loss"
	case strings.HasPrefix(n, "draw"):
		return "draw"
	}
	return ""
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func collect(args []string) []string {
	var out []string
	for _, a := range args {
		st, err := os.Stat(a)
		if err != nil {
			continue
		}
		if !st.IsDir() {
			out = append(out, a)
			continue
		}
		_ = filepath.Walk(a, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				switch strings.ToLower(filepath.Ext(p)) {
				case ".png", ".jpg", ".jpeg":
					out = append(out, p)
				}
			}
			return nil
		})
	}
	sort.Strings(out)
	return out
}

func decode(path string) (img image.Image, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return owtracker.DecodeImagePublic(f)
}
