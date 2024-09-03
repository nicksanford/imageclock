package main

import (
	"fmt"
	"image"
	"image/color"
	"sync/atomic"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

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
