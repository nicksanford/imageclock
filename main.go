package main

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/utils"

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
