package clockdrawer

import (
	"fmt"
	"image"
	"image/color"
	"sync/atomic"
	"time"

	_ "embed"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	//go:embed fonts/Aileron-Regular.otf
	fontBytes []byte
)

type ClockDrawer struct {
	Name      string
	Format    string
	Big       bool
	StartTime time.Time
	r         image.Rectangle
	face      font.Face
	color     color.Color
	count     atomic.Uint64
}

func New(
	name string,
	color color.Color,
	format string,
	big bool,
) (ClockDrawer, error) {
	if format != "jpeg" && format != "png" {
		return ClockDrawer{}, fmt.Errorf("unsupported format %s. supported formats: jpeg png", format)
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
		return ClockDrawer{}, fmt.Errorf("failed to parse font: %v", err)
	}

	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size: float64(r.Bounds().Dx() / 30),
		DPI:  72,
	})

	if err != nil {
		return ClockDrawer{}, fmt.Errorf("failed to create new face: %v", err)
	}

	return ClockDrawer{
		r:         r,
		Name:      name,
		face:      face,
		color:     color,
		Format:    format,
		Big:       big,
		StartTime: time.Now(),
	}, nil
}

func (cd *ClockDrawer) Ext() string {
	if cd.Format == "jpeg" {
		return ".jpg"
	}
	return ".png"
}
func (cd *ClockDrawer) Image(time string) *image.RGBA {
	// Make a new image with a gray background
	dst := image.NewRGBA(cd.r)

	nameDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 1 * 64)},
	}
	nameDrawer.DrawString(cd.Name)

	startTimeDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 2 * 64)},
	}
	startTimeDrawer.DrawString(fmt.Sprintf("start_time: %d", cd.StartTime.Unix()))

	timeDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 3 * 64)},
	}
	timeDrawer.DrawString(time)

	countDrawer := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(cd.color),
		Face: cd.face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(dst.Bounds().Dx() / 11 * 64), Y: fixed.Int26_6(dst.Bounds().Dy() / 5 * 4 * 64)},
	}
	countDrawer.DrawString(fmt.Sprintf("count: %d", cd.count.Add(1)))

	return dst
}
