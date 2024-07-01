package ffmpeg

import (
	"errors"
	"fmt"
	// Packages
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Re implements a resampler and rescaler for audio and video frames.
// May need to extend it for subtitles later on
type Re struct {
	t     Type
	audio *resampler
	video *rescaler
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewRe(par *Par, force bool) (*Re, error) {
	re := new(Re)
	re.t = par.Type()
	switch re.t {
	case AUDIO:
		if audio, err := NewResampler(par, force); err != nil {
			return nil, err
		} else {
			re.audio = audio
		}
	case VIDEO:
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

func (re *Re) Frame(src *Frame) (*Frame, error) {
	// Check type - if not flush
	if src != nil {
		if src.Type() != re.t {
			return nil, fmt.Errorf("frame type mismatch: %v", src.Type())
		}
	}
	switch re.t {
	case AUDIO:
		return re.audio.Frame(src)
	case VIDEO:
		return re.video.Frame(src)
	default:
		return src, nil
	}
}
