// owdump extracts training samples from a recorded stream: the banner zone of every
// frame (fps per second), resized to 128x40 RGB, appended to a raw file
// (uint8, frame-major) with a CSV index of video timestamps.
//
//	go run ./cmd/owdump -video stream.mp4 -out dataset/frames
package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/owtracker"
	"github.com/nfnt/resize"
)

const (
	srcW, srcH = 960, 540 // like a mid-size OBS frame
	outW, outH = 128, 40
)

func main() {
	video := flag.String("video", "", "recorded stream")
	ffmpeg := flag.String("ffmpeg", "ffmpeg", "ffmpeg executable")
	out := flag.String("out", "frames", "output prefix (.bin + .csv)")
	fps := flag.Float64("fps", 2, "frames per second")
	flag.Parse()
	if *video == "" {
		flag.Usage()
		os.Exit(2)
	}
	bin, err := os.Create(*out + ".bin")
	if err != nil {
		log.Fatal(err)
	}
	defer bin.Close()
	idx, err := os.Create(*out + ".csv")
	if err != nil {
		log.Fatal(err)
	}
	defer idx.Close()
	fmt.Fprintln(idx, "i,t")
	bw := bufio.NewWriterSize(bin, 1<<20)
	defer bw.Flush()

	cmd := exec.Command(*ffmpeg, "-v", "error", "-hwaccel", "auto", "-i", *video,
		"-vf", fmt.Sprintf("fps=%g,scale=%d:%d", *fps, srcW, srcH),
		"-f", "rawvideo", "-pix_fmt", "rgba", "-")
	cmd.Stderr = os.Stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	rd := bufio.NewReaderSize(pipe, 1<<22)
	buf := make([]byte, srcW*srcH*4)
	row := make([]byte, outW*outH*3)
	n := 0
	for {
		if _, err := io.ReadFull(rd, buf); err != nil {
			break
		}
		frame := &image.RGBA{Pix: buf, Stride: srcW * 4, Rect: image.Rect(0, 0, srcW, srcH)}
		small := resize.Resize(outW, outH, owtracker.CropBannerZone(frame), resize.Bilinear)
		k := 0
		for y := 0; y < outH; y++ {
			for x := 0; x < outW; x++ {
				c := color.RGBAModel.Convert(small.At(x, y)).(color.RGBA)
				row[k], row[k+1], row[k+2] = c.R, c.G, c.B
				k += 3
			}
		}
		if _, err := bw.Write(row); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(idx, "%d,%.3f\n", n, float64(n)/(*fps))
		n++
		if n%3600 == 0 {
			log.Printf("%d frames (%.0f min)", n, float64(n)/(*fps)/60)
		}
	}
	_ = cmd.Wait()
	log.Printf("done: %d frames", n)
}
