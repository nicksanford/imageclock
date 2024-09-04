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
	"slices"
	"strings"
	"time"

	"github.com/nicksanford/imageclock/clockdrawer"
	"golang.org/x/exp/maps"

	"go.viam.com/rdk/logging"
	"go.viam.com/utils"
)

func init() {
	slices.Sort(colorOptions)
}

func main() {
	utils.ContextualMain(realMain, logging.NewLogger("imageclock"))
}

func run(ctx context.Context, a []string, logger logging.Logger) error {
	return nil
}

func realMain(ctx context.Context, a []string, logger logging.Logger) error {
	if len(a) >= 2 && a[1] == "run" {
		return run(ctx, a, logger)
	}
	args, err := parseArgs(a)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(args.basepath, 0o700); err != nil {
		return fmt.Errorf("failed to create basepath directory: %v", err)
	}

	d, err := clockdrawer.New(a[0], args.color, args.format, args.big)
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
	color    color.NRGBA
	interval time.Duration
	format   string
	big      bool
}

var colors = map[string]color.NRGBA{
	"white": {R: 255, G: 255, B: 255, A: 255},
	"red":   {R: 255, A: 255},
	"green": {G: 255, A: 255},
	"blue":  {B: 255, A: 255},
}

var colorOptions = maps.Keys(colors)

func parseArgs(a []string) (args, error) {
	if len(a) != 6 {
		return args{}, fmt.Errorf("usage: %s basepath color interval format size", a[0])
	}
	basepath := a[1]

	c, ok := colors[a[2]]
	if !ok {
		return args{}, fmt.Errorf("unsupported color %s, color options: %s", a[2], strings.Join(colorOptions, " "))
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

func writeImage(cd *clockdrawer.ClockDrawer, basepath string) error {
	nowStr := time.Now().Format(time.RFC3339Nano)
	image, err := cd.Image("time: " + nowStr)
	if err != nil {
		return fmt.Errorf("failed to create image: %v", err)
	}

	f, err := os.Create(path.Join(basepath, nowStr+cd.Ext()))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer f.Close()

	b := bufio.NewWriter(f)
	if cd.Format == "jpeg" {
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
