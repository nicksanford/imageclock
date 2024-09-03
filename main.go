package main

import (
	"bufio"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	_ "embed"
)

var (
	//go:embed fonts/Aileron-Regular.otf
	fontBytes []byte
)

type clockDrawer struct {
	r        image.Rectangle
	basepath string
	face     font.Face
	color    color.Color
	interval time.Duration
	format   string
	big      bool
}

func main() {
	realMain()
}

func realMain() {
	if len(os.Args) != 6 {
		log.Fatalf("usage: %s basepath color interval format size", os.Args[0])
	}
	format := os.Args[4]

	if format != "jpeg" && format != "png" {
		log.Fatalf("unsupported format %s. supported formats: jpeg png", os.Args[4])
	}

	size := os.Args[5]
	if size != "big" && size != "small" {
		log.Fatalf("size_kb is not a number format %s. supported formats: jpeg png", os.Args[4])
	}
	big := size == "big"

	basepath := os.Args[1]
	var c color.Color
	switch os.Args[2] {
	case "white":
		c = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case "red":
		c = color.NRGBA{R: 255, A: 255}
	case "green":
		c = color.NRGBA{G: 255, A: 255}
	case "blue":
		c = color.NRGBA{B: 255, A: 255}
	default:
		log.Fatalf("unsupported color %s", os.Args[2])
	}
	interval, err := time.ParseDuration(os.Args[3])
	if err != nil {
		log.Fatalf("invalid interval: %v", err)
	}

	if err := os.MkdirAll(basepath, 0o700); err != nil {
		log.Fatalf("failed to create basepath directory: %v", err)
	}

	d := newClockDrawer(basepath, c, interval, format, big)

	log.Printf("logging %s images to %s	every %s\n", format, basepath, interval)
	for {
		newImage(d)
		time.Sleep(d.interval)
	}
}

func newClockDrawer(
	basepath string,
	color color.Color,
	interval time.Duration,
	format string,
	big bool,
) clockDrawer {
	multiple := 1
	if big {
		if format == "jpeg" {
			multiple = 4
		}
		if format == "png" {
			multiple = 8
		}
	}
	r := image.Rect(0, 0, 2560*multiple, 1440*multiple)
	// create the font
	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
	}

	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size: float64(r.Bounds().Dx() / 30),
		DPI:  72,
	})

	if err != nil {
		log.Fatalf("failed to create new face: %v", err)
	}

	return clockDrawer{
		r:        r,
		basepath: basepath,
		face:     face,
		color:    color,
		interval: interval,
		format:   format,
		big:      big,
	}

}

func newImage(cd clockDrawer) {
	// Make a new image with a gray background
	dst := image.NewRGBA(cd.r)

	// create the drawer
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 3 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 3 * 64)},
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339Nano)
	d.DrawString(nowStr)
	ext := ".png"
	if cd.format == "jpeg" {
		ext = ".jpg"
	}

	f, err := os.Create(path.Join(cd.basepath, nowStr+ext))
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer f.Close()

	b := bufio.NewWriter(f)
	if cd.format == "jpeg" {
		if err := jpeg.Encode(b, dst, &jpeg.Options{Quality: 100}); err != nil {
			log.Fatalf("failed to encode image: %v", err)
		}
	} else {
		if err := png.Encode(b, dst); err != nil {
			log.Fatalf("failed to encode image: %v", err)
		}

	}
}
