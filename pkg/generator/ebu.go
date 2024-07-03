package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"time"

	// Packages
	draw2d "github.com/llgcode/draw2d"
	draw2dimg "github.com/llgcode/draw2d/draw2dimg"
	media "github.com/mutablelogic/go-media"
	fonts "github.com/mutablelogic/go-media/etc/fonts"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type ebu struct {
	frame *ffmpeg.Frame
	image image.Image
	gc    *draw2dimg.GraphicContext
	re    *ffmpeg.Re
}

var _ Generator = (*ebu)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Set the font cache
	draw2d.SetFontCache(fonts.NewFontCache())
}

// Create a new video generator which generates the EBU Test Card
func NewEBU(par *ffmpeg.Par) (*ebu, error) {
	ebu := new(ebu)

	// Check parameters
	if par.Type() != media.VIDEO {
		return nil, errors.New("invalid codec type")
	}
	framerate := ff.AVUtil_rational_q2d(par.Framerate())
	if framerate <= 0 {
		return nil, errors.New("invalid framerate")
	}

	// Create a rescalar
	if re, err := ffmpeg.NewRe(par, false); err != nil {
		return nil, err
	} else {
		ebu.re = re
	}

	// Create a destination frame
	frame, err := ffmpeg.NewFrame(par)
	if err != nil {
		return nil, errors.Join(err, ebu.Close())
	} else {
		ebu.frame = frame
	}

	// Allocate buffer
	if err := frame.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, ebu.Close())
	}

	// Create an image for drawing
	img := image.NewRGBA(image.Rect(0, 0, frame.Width(), frame.Height()))
	if img == nil {
		return nil, errors.Join(errors.New("failed to create image"), ebu.Close())
	} else {
		ebu.image = img
	}

	// Create a graphic context for drawing
	ebu.gc = draw2dimg.NewGraphicContext(img)
	ebu.gc.SetFontData(draw2d.FontData{
		Name:   "IBMPlex",
		Family: draw2d.FontFamilyMono,
		Style:  draw2d.FontStyleNormal,
	})

	// Return success
	return ebu, nil
}

// Free resources for the generator
func (ebu *ebu) Close() error {
	var result error

	// Close resources
	if ebu.re != nil {
		result = errors.Join(result, ebu.re.Close())
	}
	if ebu.frame != nil {
		result = errors.Join(result, ebu.frame.Close())
	}

	// Release resources
	ebu.re = nil
	ebu.frame = nil
	ebu.image = nil
	ebu.gc = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ebu *ebu) String() string {
	data, _ := json.MarshalIndent(ebu.frame, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the first and subsequent frames of video data
func (ebu *ebu) Frame() *ffmpeg.Frame {
	// Set the Pts
	if ebu.frame.Pts() == ffmpeg.PTS_UNDEFINED {
		ebu.frame.SetPts(0)
	} else {
		ebu.frame.IncPts(1)
	}

	// Calculcate the width/height
	w, h := float64(ebu.frame.Width()), float64(ebu.frame.Height())

	// Draw the EBU test card with overlayed timestamp
	draw_bands(ebu.gc, w, h)
	draw_text(ebu.gc, w, h, ebu.Ts(), w/20)

	// Copy image to frame
	if err := ebu.frame.FromImage(ebu.image); err != nil {
		return nil
	}

	// Resize the frame and return it
	frame, err := ebu.re.Frame(ebu.frame)
	if err != nil {
		return nil
	} else {
		return frame
	}
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// 10-bit YCbCr values for the eight standard colors.
type YCbCr10 struct {
	Y, Cb, Cr uint16
}

var (
	White   = YCbCr10{940, 512, 512}
	Yellow  = YCbCr10{877, 64, 553}
	Cyan    = YCbCr10{754, 615, 64}
	Green   = YCbCr10{691, 167, 105}
	Magenta = YCbCr10{313, 857, 919}
	Red     = YCbCr10{250, 409, 960}
	Blue    = YCbCr10{127, 960, 471}
	Black   = YCbCr10{64, 512, 512}
)

const (
	nullTimestamp = "--:--:--.--"
)

// Make 8-bit
func (c YCbCr10) Color() color.YCbCr {
	return color.YCbCr{
		Y:  uint8(c.Y >> 2),
		Cb: uint8(c.Cb >> 2),
		Cr: uint8(c.Cr >> 2),
	}
}

// Size timestamp box
func (ebu *ebu) Ts() string {
	// Convert the Ts to "00:00:00.00"
	ts := ebu.frame.Ts()
	if ts < 0 {
		return nullTimestamp
	}
	duration := time.Duration(ts * float64(time.Second))
	f := ebu.frame.Pts() % int64(ebu.frame.TimeBase().Den())
	return fmt.Sprintf("%02d:%02d:%02d.%02d", int(duration.Hours())%99, int(duration.Minutes())%60, int(duration.Seconds())%60, f)
}

func draw_bands(gc draw2d.GraphicContext, w, h float64) {
	colors := []YCbCr10{White, Yellow, Cyan, Green, Magenta, Red, Blue, Black}
	band_width := w / float64(len(colors))
	for i := 0; i < len(colors); i++ {
		x1, y1 := float64(i)*band_width, float64(0)
		x2, y2 := float64((i+1))*band_width, h

		// Draw a rect
		gc.SetFillColor(colors[i].Color())
		gc.BeginPath()
		gc.MoveTo(x1, y1)
		gc.LineTo(x2, y1)
		gc.LineTo(x2, y2)
		gc.LineTo(x1, y2)
		gc.Close()
		gc.Fill()
	}
}

func draw_text(gc draw2d.GraphicContext, w, h float64, text string, size float64) {
	gc.SetFontSize(size)
	left, top, right, bottom := gc.GetStringBounds(nullTimestamp)

	x1, y1 := (w/2)-(right-left)/2, (h+bottom-top)/2
	x2, y2 := x1+right-left, y1+bottom-top
	border := size / 2

	gc.SetFillColor(Black.Color())
	gc.BeginPath()
	gc.MoveTo(x1-border, y1-border)
	gc.LineTo(x2+border, y1-border)
	gc.LineTo(x2+border, y2+border)
	gc.LineTo(x1-border, y2+border)
	gc.Close()
	gc.Fill()

	gc.SetFillColor(White.Color())
	gc.FillStringAt(text, x1, y1+bottom-top)
}
