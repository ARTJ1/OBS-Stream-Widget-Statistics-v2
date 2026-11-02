// owrank runs the auto rank reader (alpha) over a recorded stream, like the live
// widget: the banner CNN counts a match, then the post-match progress screen is
// read for the role and rank. It prints every counted match and the rank it read.
//
//	go run ./cmd/owrank -video match.mp4 [-fps 2] [-all] [-thumbs dir]
//
// -all reads every frame (no banner needed) to list every progress screen in the
// video: the ground truth to compare the production mode against.
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
	thumbs := flag.String("thumbs", "", "save a JPEG of every frame with a role read here")
	hwaccel := flag.String("hwaccel", "auto", "ffmpeg -hwaccel (none = CPU decoding, faster for several videos at once)")
	noExpect := flag.Bool("noexpect", false, "ignore the rank before the match (the old reader)")
	realtime := flag.Bool("realtime", false, "read frames at the live widget's pace: banner every 250 ms, rank as the live reader (use -fps 20)")
	watch := flag.String("watch", "", "save one JPEG per second of every post-match watch here (one folder per match)")
	noRank := flag.Bool("norank", false, "count matches only (fast: no rank reader)")
	banner := flag.Bool("banner", false, "print the banner CNN's win/loss odds for every frame that looks like a banner")
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
	var watchDir string
	var busyUntil, nextBanner time.Duration

	args := []string{"-v", "error", "-hwaccel", *hwaccel}
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
	var matches, ranks, ocrN int
	var ocrTime time.Duration
	var countedAt time.Duration
	// Like the widget: the last rank read per role is the next match's prior.
	known := map[string]int{}
	prior := func(role string) (int, bool) { r, ok := known[role]; return r, ok }
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
			if *realtime && vt < busyUntil {
				continue // the live reader is still busy with the last frame
			}
			t0 := time.Now()
			r, conf, ok, err := rs.Frame(frame, at)
			took := time.Since(t0)
			ocrTime += took
			ocrN++
			busyUntil = vt + took + rs.Delay(r, at, took) // the live reader's pace
			if err != nil {
				log.Fatalf("ocr: %v", err)
			}
			if r.Role != "" {
				fmt.Printf("%s   read role=%-7s rank=%-14s label=%q\n", clock(vt), r.Role, rankName(r), r.RankText)
				saveThumb(*thumbs, vt, frame)
			}
			if watchDir != "" && (r.Role != "" || i%int(*fps) == 0) {
				saveThumb(watchDir, vt, frame)
			}
			if ok {
				ranks++
				known[conf.Role] = conf.Rank
				fmt.Printf("%s RANK %s %s (+%s after the banner)\n", clock(vt), conf.Role, rankName(conf), (vt - countedAt).Round(time.Second))
			} else if !rs.Active() {
				fmt.Printf("%s RANK none (watch ended)\n", clock(vt))
			}
			continue
		}
		if cs.Cooling(at) {
			continue
		}
		if *realtime {
			if vt < nextBanner {
				continue
			}
			nextBanner = vt + 250*time.Millisecond // the live banner poll
		}
		small := resize.Resize(480, 270, frame, resize.Bilinear)
		br, outcome, counted := cs.Frame(small, at)
		if *banner && (br.WinP >= 0.1 || br.LossP >= 0.1) {
			fmt.Printf("%s.%03d banner win=%.2f loss=%.2f vote=%q\n", clock(vt), vt.Milliseconds()%1000, br.WinP, br.LossP, br.Outcome)
		}
		if counted {
			matches++
			countedAt = vt
			fmt.Printf("%s COUNTED %s\n", clock(vt), outcome)
			expect := owtracker.RankExpect(outcome, prior)
			if *noExpect {
				expect = nil
			}
			if !*noRank {
				rs.Start(at, expect)
			}
			if *watch != "" {
				s := int(vt.Seconds())
				watchDir = filepath.Join(*watch, fmt.Sprintf("%dh%02dm%02ds_%s", s/3600, s/60%60, s%60, outcome))
				_ = os.MkdirAll(watchDir, 0o755)
			}
		}
	}
	_ = cmd.Wait()
	avg := time.Duration(0)
	if ocrN > 0 {
		avg = ocrTime / time.Duration(ocrN)
	}
	fmt.Printf("\nscanned in %s: %d matches, %d ranks read, %d reader calls (avg %s)\n",
		time.Since(began).Round(time.Second), matches, ranks, ocrN, avg.Round(time.Millisecond))
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
