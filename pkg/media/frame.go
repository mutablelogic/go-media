package media

import (

	// Packages
	"fmt"

	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type frame struct {
	ctx *ffmpeg.AVFrame
}

// Ensure *input complies with Media interface
var _ Frame = (*frame)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFrame() *frame {
	frame := new(frame)

	if ctx := ffmpeg.AVUtil_frame_alloc(); ctx == nil {
		return nil
	} else {
		frame.ctx = ctx
	}

	// Return success
	return frame
}

func (frame *frame) Close() error {
	var result error

	// Callback
	if frame.ctx != nil {
		ffmpeg.AVUtil_frame_free_ptr(frame.ctx)
		frame.ctx = nil
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (frame *frame) String() string {
	str := "<media.frame"
	if frame.ctx != nil {
		str += fmt.Sprint(" ctx=", frame.ctx)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Unref releases the packet
func (frame *frame) Release() {
	if frame.ctx != nil {
		ffmpeg.AVUtil_frame_unref(frame.ctx)
	}
}
