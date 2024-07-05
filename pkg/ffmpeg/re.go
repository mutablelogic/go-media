package ffmpeg

import (
	"errors"
	"fmt"

	// Packages
	media "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Re implements a resampler and rescaler for audio and video frames.
// May need to extend it for subtitles later on
type Re struct {
	t     media.Type
	audio *resampler
	video *rescaler
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return a new resampler or rescaler, with the destination parameters. If
// force is true, then the resampler will always resample, even if the
// destination parameters are the same as the source parameters.
func NewRe(par *Par, force bool) (*Re, error) {
	re := new(Re)
	re.t = par.Type()
	switch re.t {
	case media.AUDIO:
		if audio, err := NewResampler(par, force); err != nil {
			return nil, err
		} else {
			re.audio = audio
		}
	case media.VIDEO:
		if video, err := NewRescaler(par, force); err != nil {
			return nil, err
		} else {
			re.video = video
		}
	default:
		return nil, fmt.Errorf("invalid resampling/rescaling type: %v", par.Type())
	}

	// Return success
	return re, nil
}

// Release resources
func (re *Re) Close() error {
	var result error
	if re.audio != nil {
		result = errors.Join(result, re.audio.Close())
	}
	if re.video != nil {
		result = errors.Join(result, re.video.Close())
	}
	re.audio = nil
	re.video = nil
	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Resample or rescale the source frame and return the destination frame
func (re *Re) Frame(src *Frame) (*Frame, error) {
	// Check type - if not flush
	if src != nil {
		if src.Type() != re.t {
			return nil, fmt.Errorf("frame type mismatch: %v", src.Type())
		}
	}
	switch re.t {
	case media.AUDIO:
		return re.audio.Frame(src)
	case media.VIDEO:
		return re.video.Frame(src)
	default:
		return src, nil
	}
}
