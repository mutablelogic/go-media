package chromaprint

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"

	// Package imports
	"github.com/mutablelogic/go-media/sys/chromaprint"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Stream struct {
	*chromaprint.Context
	AudioFormat

	channels          int
	duration          float64
	shifts_per_sample int
	samples_per_sec   float64
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new fingerprint stream with audio format, channel layout and sample rate in Hz
func NewStream(f AudioFormat, c AudioChannelLayout, r uint) (*Stream, error) {
	stream := new(Stream)

	// Check audio format
	if b := shiftsPerSample(f); b < 0 {
		return nil, ErrNotImplemented.With(f)
	} else {
		stream.AudioFormat = f
		stream.shifts_per_sample = b
	}

	// Check channels and rate
	if c.Channels == 0 || r == 0 {
		return nil, ErrBadParameter.With("NewStream")
	} else {
		stream.samples_per_sec = float64(r)
		stream.channels = int(c.Channels)
	}

	// Set context, rate and channels
	ctx := chromaprint.NewChromaprint(chromaprint.ALGORITHM_DEFAULT)
	if ctx == nil {
		return nil, ErrInternalAppError.With("NewStream")
	}

	// Start fingerprint process
	if err := ctx.Start(int(r), int(c.Channels)); err != nil {
		ctx.Free()
		return nil, err
	} else {
		stream.Context = ctx
	}

	// Return success
	return stream, nil
}

// Release stream resources
func (stream *Stream) Release() {
	if stream.Context != nil {
		stream.Context.Free()
		stream.Context = nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (stream *Stream) String() string {
	str := "<chromaprint"
	if v := chromaprint.Version(); v != "" {
		str += fmt.Sprintf(" version=%q", v)
	}
	if f := stream.AudioFormat; f != 0 {
		str += fmt.Sprint(" format=", f)
	}
	if d := stream.Duration(); d != 0 {
		str += fmt.Sprint(" duration=", d)
	}
	if c := stream.Channels(); c != 0 {
		str += fmt.Sprint(" internal_ch=", c)
	}
	if r := stream.Rate(); r != 0 {
		str += fmt.Sprintf(" internal_rate=%dHz", r)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get the sampling rate that is internally used for fingerprinting
func (stream *Stream) Rate() uint {
	if stream.Context == nil {
		return 0
	}
	return uint(stream.Context.Rate())
}

// Get the number of channels that is internally used for fingerprinting
func (stream *Stream) Channels() uint {
	if stream.Context == nil {
		return 0
	}
	return uint(stream.Context.Channels())
}

// Get the current duration
func (stream *Stream) Duration() time.Duration {
	if stream.Context == nil {
		return 0
	} else {
		return time.Duration(stream.duration * float64(time.Second))
	}
}

// Used for all types of stream
func (stream *Stream) Write(data []byte) error {
	if stream.Context == nil {
		return ErrOutOfOrder.With("WriteUint8")
	}

	// Calculate number of samples and hence add duration
	samples := len(data) >> stream.shifts_per_sample

	// Write data
	fmt.Println("data=", strings.ToUpper(hex.EncodeToString(data)), " len=", len(data), " samples=", samples)
	if err := stream.Context.Write(data); err != nil {
		return err
	}

	// Calculate number of samples and hence add duration
	stream.duration += float64(samples) / float64(stream.samples_per_sec)

	// Return success
	return nil
}

// Return fingerprint
func (stream *Stream) Fingerprint() (string, error) {
	if stream.Context == nil {
		return "", ErrOutOfOrder.With("Fingerprint")
	}
	if stream.duration == 0 {
		return "", ErrOutOfOrder.With("Fingerprint")
	}
	err := stream.Finish()
	if err != nil {
		return "", err
	}
	fp, err := stream.GetFingerprint()
	if err != nil {
		return "", err
	}

	// Return success
	return fp, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func shiftsPerSample(format AudioFormat) int {
	switch format {
	case AUDIO_FMT_U8:
		return 0
	case AUDIO_FMT_S16:
		return 1
	case AUDIO_FMT_S32, AUDIO_FMT_F32:
		return 2
	case AUDIO_FMT_S64, AUDIO_FMT_F64:
		return 3
	}
	// Returns -1 otherwise
	return -1
}
