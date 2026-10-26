// owrank runs the auto rank reader (alpha) over a recorded stream, like the live
// widget: the banner CNN counts a match, then the post-match progress screen is
// read for the role and rank. It prints every counted match and the rank it read.
//
//	go run ./cmd/owrank -video match.mp4 [-fps 2] [-all] [-thumbs dir]
//
// -all reads every frame (no banner needed) to list every progress screen in the
// video: the ground truth to compare the production mode against.
//
// -gate also runs the live fallback: outside the post-match watch, a cheap check
// on the small frame opens a short watch when it sees the progress screen (draws,
// missed banners).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
	"github.com/nfnt/resize"
)

// Live widget: OBS frames at 1280 px for the rank reader, 480 px for the banner CNN.
const frameW, frameH = 1280, 720

func main() {
	video := flag.String("video", "", "recorded stream (mp4)")
	ffmpeg := flag.String("ffmpeg", "ffmpeg", "ffmpeg executable")
	fps := flag.Float64("fps", 2, "frames per second")
	start := flag.Duration("start", 0, "skip to this video time")
	limit := flag.Duration("limit", 0, "scan only this much video (0 = all)")
	all := flag.Bool("all", false, "ground truth: read every frame, no banner needed")
	gate := flag.Bool("gate", false, "also open a watch from the cheap progress-screen check (live fallback)")
	thumbs := flag.String("thumbs", "", "save a JPEG of every frame with a role read here")
	flag.Parse()
	if *video == "" {
		flag.Usage()
		os.Exit(2)
	}
	work, err := os.MkdirTemp("", "owrank")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(work)
	rs := owtracker.NewRankScanner(work)
	defer rs.Close()
	cs, err := owtracker.NewCNNScanner("")
	if err != nil {
		log.Fatal(err)
	}
	if *thumbs != "" {
		_ = os.MkdirAll(*thumbs, 0o755)
	}

	args := []string{"-v", "error", "-hwaccel", "auto"}
	if *start > 0 {
		args = append(args, "-ss", fmt.Sprint(start.Seconds()))
	}
	args = append(args, "-i", *video)
	if *limit > 0 {
		args = append(args, "-t", fmt.Sprint(limit.Seconds()))
	}
	args = append(args, "-vf", fmt.Sprintf("fps=%g,scale=%d:%d", *fps, frameW, frameH),
		"-f", "rawvideo", "-pix_fmt", "rgba", "-")
	cmd := exec.Command(*ffmpeg, args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	rd := bufio.NewReaderSize(out, 1<<22)

	base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	buf := make([]byte, frameW*frameH*4)
	var matches, ranks, ocrN, gateN int
	var gateQuiet time.Time
	byGate := false
	var ocrTime time.Duration
	var countedAt time.Duration
	began := time.Now()
	for i := 0; ; i++ {
		if _, err := io.ReadFull(rd, buf); err != nil {
			break
		}
		vt := *start + time.Duration(float64(i)/(*fps)*float64(time.Second))
		at := base.Add(vt)
		frame := &image.RGBA{Pix: buf, Stride: frameW * 4, Rect: image.Rect(0, 0, frameW, frameH)}

		if *all {
			t0 := time.Now()
			r, err := rs.Read(frame)
			ocrTime += time.Since(t0)
			ocrN++
			if err != nil {
				log.Fatalf("ocr: %v", err)
			}
			if r.Role != "" {
				fmt.Printf("%s READ role=%-7s rank=%-14s role=%q label=%q\n", clock(vt), r.Role, rankName(r), r.RoleText, r.RankText)
				saveThumb(*thumbs, vt, frame)
			}
			continue
		}

		if rs.Active() {
			t0 := time.Now()
			r, conf, ok, err := rs.Frame(frame, at)
			ocrTime += time.Since(t0)
			ocrN++
			if err != nil {
				log.Fatalf("ocr: %v", err)
			}
			if r.Role != "" {
				fmt.Printf("%s   read role=%-7s rank=%-14s label=%q\n", clock(vt), r.Role, rankName(r), r.RankText)
				saveThumb(*thumbs, vt, frame)
			}
			if ok {
				ranks++
				fmt.Printf("%s RANK %s %s (+%s after the banner)\n", clock(vt), conf.Role, rankName(conf), (vt - countedAt).Round(time.Second))
			} else if !rs.Active() {
				fmt.Printf("%s RANK none (watch ended)\n", clock(vt))
			}
			if !rs.Active() && byGate {
				byGate, gateQuiet = false, at.Add(60*time.Second)
			}
			continue
		}
		if cs.Cooling(at) {
			continue
		}
		small := resize.Resize(480, 270, frame, resize.Bilinear)
		if _, outcome, counted := cs.Frame(small, at); counted {
			matches++
			countedAt = vt
			fmt.Printf("%s COUNTED %s\n", clock(vt), outcome)
			rs.Start(at)
		} else if *gate && !at.Before(gateQuiet) && rs.Gate(small) {
			gateN++
			t0 := time.Now()
			r, err := rs.Read(frame)
			ocrTime += time.Since(t0)
			ocrN++
			if err != nil {
				log.Fatalf("ocr: %v", err)
			}
			if r.Role != "" {
				// Live: the watch opens and this read is its first vote.
				fmt.Printf("%s GATE role=%-7s rank=%-14s label=%q\n", clock(vt), r.Role, rankName(r), r.RankText)
				countedAt, byGate = vt, true
				rs.StartGate(at)
				rs.Observe(at, r)
			}
		}
	}
	_ = cmd.Wait()
	avg := time.Duration(0)
	if ocrN > 0 {
		avg = ocrTime / time.Duration(ocrN)
	}
	fmt.Printf("\nscanned in %s: %d matches, %d ranks read, %d reader calls (avg %s), %d gate passes\n",
		time.Since(began).Round(time.Second), matches, ranks, ocrN, avg.Round(time.Millisecond), gateN)
}

func rankName(r owtracker.RankRead) string {
	if r.Tier == "" {
		return "-"
	}
	return fmt.Sprintf("%s %d", r.Tier, r.Division)
}

func saveThumb(dir string, vt time.Duration, img image.Image) {
	if dir == "" {
		return
	}
	s := int(vt.Seconds())
	f, err := os.Create(filepath.Join(dir, fmt.Sprintf("%dh%02dm%02ds_%03d.jpg", s/3600, s/60%60, s%60, vt.Milliseconds()%1000)))
	if err != nil {
		return
	}
	defer f.Close()
	_ = jpeg.Encode(f, img, &jpeg.Options{Quality: 70})
}

func clock(d time.Duration) string {
	s := int(d.Seconds())
	return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
}
