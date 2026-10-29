package thumbnailer

import (
	// Packages
	gomedia "github.com/mutablelogic/go-media"
	reader "github.com/mutablelogic/go-media/reader"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a function which applies options to a Thumbnailer
type Opt func(*opts) error

// Mode is how frames are sampled from a video
type Mode int

type opts struct {
	width, height int          // size of thumbnails, either of which can be zero to keep the aspect ratio
	mode          Mode         // how frames are sampled
	maxSamples    int          // maximum number of frames sampled by seeking
	stream        int          // video stream, or -1 for the first one
	readerOpts    []reader.Opt // reader options
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// How frames are sampled
const (
	ModeAuto      Mode = iota // Keyframes if the video has them at a useful rate, otherwise seek
	ModeKeyframes             // Decode only the keyframes
	ModeSeek                  // Seek to evenly spaced times, and decode the frame at each
)

const (
	DefaultWidth      = 640 // Default width of thumbnails
	DefaultMaxSamples = 300 // Default maximum number of frames sampled by seeking
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opt ...Opt) (*opts, error) {
	o := opts{width: DefaultWidth, maxSamples: DefaultMaxSamples, stream: -1}
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}
	return &o, nil
}

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// WithSize sets the size of thumbnails. Either can be zero, to keep the
// video's aspect ratio, and both zero is the video's own size. The default is
// DefaultWidth wide.
func WithSize(width, height int) Opt {
	return func(o *opts) error {
		if width < 0 || height < 0 {
			return gomedia.ErrBadParameter.Withf("invalid size %dx%d", width, height)
		}
		o.width, o.height = width, height
		return nil
	}
}

// WithMode sets how frames are sampled from the video.
func WithMode(mode Mode) Opt {
	return func(o *opts) error {
		if mode < ModeAuto || mode > ModeSeek {
			return gomedia.ErrBadParameter.Withf("invalid mode %v", mode)
		}
		o.mode = mode
		return nil
	}
}

// WithMaxSamples sets the maximum number of frames sampled by seeking, which
// sets how far apart they are. The default is DefaultMaxSamples.
func WithMaxSamples(n int) Opt {
	return func(o *opts) error {
		if n < 1 {
			return gomedia.ErrBadParameter.Withf("invalid maximum number of samples %d", n)
		}
		o.maxSamples = n
		return nil
	}
}

// WithStream sets the index of the video stream. Without it, the video
// stream with the lowest index is used.
func WithStream(stream int) Opt {
	return func(o *opts) error {
		if stream < 0 {
			return gomedia.ErrBadParameter.Withf("invalid stream %d", stream)
		}
		o.stream = stream
		return nil
	}
}

// WithReaderOpt passes a reader.Opt to the reader.
func WithReaderOpt(opt reader.Opt) Opt {
	return func(o *opts) error {
		o.readerOpts = append(o.readerOpts, opt)
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m Mode) String() string {
	switch m {
	case ModeAuto:
		return "ModeAuto"
	case ModeKeyframes:
		return "ModeKeyframes"
	case ModeSeek:
		return "ModeSeek"
	default:
		return "Mode(?)"
	}
}
