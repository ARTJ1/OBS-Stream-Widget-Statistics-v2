// owscan runs the auto win/loss detector over a recorded stream, exactly like the
// live widget with OBS frames: a 320 px frame per second (stage 1) and a full-size
// frame only when the strip fires (stage 2, OCR). It prints every counted result
// with its video time so it can be checked against what really happened.
//
//	go run ./cmd/owscan -video match.mp4 [-ffmpeg path] [-fps 1] [-thumbs dir]
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
)

const probeW, probeH = 320, 180

func main() {
	video := flag.String("video", "", "recorded stream (mp4)")
	ffmpeg := flag.String("ffmpeg", "ffmpeg", "ffmpeg executable")
	fps := flag.Float64("fps", 1, "stage-1 frames per second (live widget: 1)")
	start := flag.Duration("start", 0, "skip to this video time")
	limit := flag.Duration("limit", 0, "scan only this much video (0 = all)")
	thumbs := flag.String("thumbs", "", "save a JPEG of every stage-2 frame here (for review)")
	truth := flag.Int("truth", 0, "ground-truth mode: stage-1 threshold %% (e.g. 8), no flow, OCR every candidate")
	cnn := flag.String("cnn", "", "production mode: banner CNN (model path, or \"embedded\") at -fps (live widget: 4) on 480 px frames")
	endorse := flag.Bool("endorse", false, "ground-truth mode: list endorse screens (ПОХВАЛИТЬ / ENDORSE) (one after every match)")
	find := flag.String("find", "", "ground-truth mode: comma-separated banner words to list (e.g. VICTORY,DEFEAT)")
	findZone := flag.String("findzone", "0.25,0.15,0.50,0.40", "-find plain-OCR region x,y,w,h (fractions)")
	flag.Parse()
	var zone [4]float64
	if _, err := fmt.Sscanf(*findZone, "%g,%g,%g,%g", &zone[0], &zone[1], &zone[2], &zone[3]); err != nil {
		log.Fatalf("-findzone: %v", err)
	}
	if *video == "" {
		flag.Usage()
		os.Exit(2)
	}
	work, err := os.MkdirTemp("", "owscan")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(work)
	sc := owtracker.NewScanner(work)
	defer sc.Close()
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
	fw, fh := probeW, probeH
	if *endorse || *find != "" {
		fw, fh = 1280, 720
	}
	var cs *owtracker.CNNScanner
	if *cnn != "" {
		fw, fh = 480, 270
		mp := *cnn
		if mp == "embedded" {
			mp = ""
		}
		if cs, err = owtracker.NewCNNScanner(mp); err != nil {
			log.Fatal(err)
		}
	}
	args = append(args, "-vf", fmt.Sprintf("fps=%g,scale=%d:%d", *fps, fw, fh),
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
	rd := bufio.NewReaderSize(out, 1<<20)

	base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	frameSize := fw * fh * 4
	buf := make([]byte, frameSize)
	var hits, reads, wins, losses int
	began := time.Now()
	for i := 0; ; i++ {
		if _, err := io.ReadFull(rd, buf); err != nil {
			break
		}
		vt := *start + time.Duration(float64(i)/(*fps)*float64(time.Second))
		at := base.Add(vt)
		small := &image.RGBA{Pix: buf, Stride: fw * 4, Rect: image.Rect(0, 0, fw, fh)}
		if cs != nil {
			if cs.Cooling(at) {
				continue
			}
			r, outcome, counted := cs.Frame(small, at)
			if counted {
				fmt.Printf("%s COUNTED %s (win=%.2f loss=%.2f)\n", clock(vt), outcome, r.WinP, r.LossP)
				if outcome == owtracker.OutcomeWin {
					wins++
				} else {
					losses++
				}
			}
			continue
		}
		if *find != "" {
			// Banner reader first (mask + deskew handles the stylised words), then plain OCR.
			if res, err := sc.ReadOnly(small); err == nil && res.Outcome != "" {
				fmt.Printf("%s FIND %-8s %q\n", clockMs(vt), res.Outcome, res.Text)
				continue
			}
			if txt, err := sc.TextAt(small, zone[0], zone[1], zone[2], zone[3]); err == nil {
				up := strings.ToUpper(txt)
				for _, w := range strings.Split(strings.ToUpper(*find), ",") {
					if w != "" && strings.Contains(up, w) {
						fmt.Printf("%s FIND %-8s %q\n", clockMs(vt), w, txt)
						break
					}
				}
			}
			continue
		}
		if *endorse {
			if txt, err := sc.TextAt(small, 0.15, 0.80, 0.70, 0.12); err == nil && (strings.Contains(strings.ToUpper(txt), "ПОХВАЛ") || strings.Contains(strings.ToUpper(txt), "ENDORSE")) {
				fmt.Printf("%s ENDORSE %q\n", clock(vt), txt)
			}
			continue
		}
		if *truth > 0 {
			if hit, g, r := sc.ProbeAt(small, *truth); hit {
				if big, err := grabFrame(*ffmpeg, *video, vt); err == nil {
					if res, err := sc.ReadOnly(big); err == nil && res.Outcome != "" {
						fmt.Printf("%s TRUTH %-5s gold=%2d%% red=%2d%% text=%q\n", clock(vt), res.Outcome, g, r, res.Text)
					}
				}
			}
			continue
		}
		if sc.Cooling(at) {
			continue
		}
		hit, g, r := sc.Probe(small)
		if !hit {
			sc.Idle(at)
			continue
		}
		hits++
		big, err := grabFrame(*ffmpeg, *video, vt)
		if err != nil {
			log.Printf("%s grab: %v", clock(vt), err)
			continue
		}
		reads++
		res, outcome, counted, err := sc.Read(big, at)
		if err != nil {
			log.Fatalf("ocr: %v", err)
		}
		mark := ""
		if counted {
			mark = "  <<< COUNTED " + string(outcome)
			if outcome == owtracker.OutcomeWin {
				wins++
			} else {
				losses++
			}
		}
		fmt.Printf("%s strip gold=%2d%% red=%2d%% read=%-5s text=%-14q%s\n",
			clock(vt), g, r, orDash(string(res.Outcome)), res.Text, mark)
		if *thumbs != "" {
			saveThumb(filepath.Join(*thumbs, fmt.Sprintf("%s_%s.jpg", clockFile(vt), orDash(string(res.Outcome)))), big)
		}
	}
	_ = cmd.Wait()
	fmt.Printf("\nscanned in %s: strip hits %d, OCR reads %d, counted %d wins, %d losses\n",
		time.Since(began).Round(time.Second), hits, reads, wins, losses)
}

// grabFrame decodes one full-size frame at video time t (fast keyframe seek).
func grabFrame(ffmpeg, video string, t time.Duration) (image.Image, error) {
	cmd := exec.Command(ffmpeg, "-v", "error", "-ss", fmt.Sprintf("%.3f", t.Seconds()), "-i", video,
		"-frames:v", "1", "-f", "image2pipe", "-vcodec", "png", "-")
	var b bytes.Buffer
	cmd.Stdout = &b
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	img, _, err := image.Decode(&b)
	return img, err
}

func saveThumb(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_ = jpeg.Encode(f, img, &jpeg.Options{Quality: 70})
}

func clockMs(d time.Duration) string {
	return fmt.Sprintf("%.2f", d.Seconds())
}

func clock(d time.Duration) string {
	s := int(d.Seconds())
	return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
}

func clockFile(d time.Duration) string {
	s := int(d.Seconds())
	return fmt.Sprintf("%dh%02dm%02ds", s/3600, s/60%60, s%60)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
