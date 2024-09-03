package main

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"sync/atomic"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/utils"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	_ "embed"
)

var (
	//go:embed fonts/Aileron-Regular.otf
	fontBytes []byte
)

func main() {
	utils.ContextualMain(realMain, logging.NewLogger("imageclock"))
}

func realMain(ctx context.Context, a []string, logger logging.Logger) error {
	args, err := parseArgs(a)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(args.basepath, 0o700); err != nil {
		return fmt.Errorf("failed to create basepath directory: %v", err)
	}

	d, err := newClockDrawer(args.color, args.format, args.big)
	if err != nil {
		return err
	}

	logger.Infof("logging %s images to %s	every %s\n", args.format, args.basepath, args.interval)
	for utils.SelectContextOrWait(ctx, args.interval) {
		if err := writeImage(&d, args.basepath); err != nil {
			return err
		}
	}
	return nil
}

type args struct {
	basepath string
	color    color.Color
	interval time.Duration
	format   string
	big      bool
}

func parseArgs(a []string) (args, error) {
	if len(a) != 6 {
		return args{}, fmt.Errorf("usage: %s basepath color interval format size", a[0])
	}
	basepath := a[1]

	var c color.Color
	switch a[2] {
	case "white":
		c = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case "red":
		c = color.NRGBA{R: 255, A: 255}
	case "green":
		c = color.NRGBA{G: 255, A: 255}
	case "blue":
		c = color.NRGBA{B: 255, A: 255}
	default:
		return args{}, fmt.Errorf("unsupported color %s", a[2])
	}
	interval, err := time.ParseDuration(a[3])
	if err != nil {
		return args{}, fmt.Errorf("invalid interval: %v", err)
	}
	format := a[4]
	size := a[5]
	if size != "big" && size != "small" {
		return args{}, fmt.Errorf("size is one of the supported options %s. supported sizes: big small", a[5])
	}
	big := size == "big"

	return args{
		basepath: basepath,
		color:    c,
		interval: interval,
		format:   format,
		big:      big,
	}, nil
}

type clockDrawer struct {
	r         image.Rectangle
	name      string
	face      font.Face
	color     color.Color
	format    string
	big       bool
	startTime time.Time
	count     atomic.Uint64
}

func newClockDrawer(
	color color.Color,
	format string,
	big bool,
) (clockDrawer, error) {
	if format != "jpeg" && format != "png" {
		return clockDrawer{}, fmt.Errorf("unsupported format %s. supported formats: jpeg png", format)
	}

	multiple := 1
	if big {
		if format == "jpeg" {
			multiple = 3
		}
		if format == "png" {
			multiple = 8
		}
	}
	r := image.Rect(0, 0, 2560*multiple, 1440*multiple)
	// create the font
	parsedFont, err := opentype.Parse(fontBytes)
	if err != nil {
		return clockDrawer{}, fmt.Errorf("failed to parse font: %v", err)
	}

	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size: float64(r.Bounds().Dx() / 30),
		DPI:  72,
	})

	if err != nil {
		return clockDrawer{}, fmt.Errorf("failed to create new face: %v", err)
	}

	return clockDrawer{
		r:         r,
		face:      face,
		color:     color,
		format:    format,
		big:       big,
		startTime: time.Now(),
	}, nil
}

func (cd *clockDrawer) ext() string {
	if cd.format == "jpeg" {
		return ".jpg"
	}
	return ".png"
}
func (cd *clockDrawer) image(time string) *image.RGBA {
	// Make a new image with a gray background
	dst := image.NewRGBA(cd.r)

	nameDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 3 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 1 * 64)},
	}
	nameDrawer.DrawString(cd.name)

	startTimeDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 3 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 2 * 64)},
	}
	startTimeDrawer.DrawString(fmt.Sprintf("start_time: %d", cd.startTime.Unix()))

	timeDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 3 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 3 * 64)},
	}
	timeDrawer.DrawString(time)

	countDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 3 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 4 * 64)},
	}
	countDrawer.DrawString(fmt.Sprintf("count: %d", cd.count.Add(1)))

	return dst
}

func writeImage(cd *clockDrawer, basepath string) error {
	nowStr := time.Now().Format(time.RFC3339Nano)
	image := cd.image(nowStr)

	f, err := os.Create(path.Join(basepath, nowStr+cd.ext()))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	b := bufio.NewWriter(f)
	if cd.format == "jpeg" {
		if err := jpeg.Encode(b, image, &jpeg.Options{Quality: 100}); err != nil {
			return fmt.Errorf("failed to encode image: %v", err)
		}
	} else {
		if err := png.Encode(b, image); err != nil {
			return fmt.Errorf("failed to encode image: %v", err)
		}
	}

	return nil
}
