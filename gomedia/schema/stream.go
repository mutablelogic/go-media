package schema

import (
	"encoding/json"
	"fmt"
	"strconv"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffschema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Stream wraps pkg/ffmpeg/schema Stream and adds CLI table formatting helpers.
type Stream struct {
	*ffschema.Stream
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func WrapStream(s *ffschema.Stream) *Stream {
	if s == nil {
		return nil
	}
	return &Stream{Stream: s}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s Stream) MarshalJSON() ([]byte, error) {
	if s.Stream == nil {
		return json.Marshal(nil)
	}
	return s.Stream.MarshalJSON()
}

func (s Stream) String() string {
	if s.Stream == nil {
		return "<nil>"
	}
	return s.Stream.String()
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (Stream) Header() []string {
	return []string{"Index", "Type", "Codec", "Details"}
}

func (s Stream) Cell(col int) string {
	switch col {
	case 0:
		if s.Stream == nil {
			return ""
		}
		return strconv.Itoa(s.Index())
	case 1:
		if s.Stream == nil {
			return ""
		}
		return s.Type().String()
	case 2:
		if s.Stream == nil || s.CodecPar() == nil {
			return ""
		}
		return s.CodecPar().CodecID().Name()
	case 3:
		return s.Details()
	default:
		return ""
	}
}

func (Stream) Width(col int) int {
	switch col {
	case 0:
		return 8
	case 1:
		return 10
	case 2:
		return 20
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s Stream) Details() string {
	if s.Stream == nil || s.CodecPar() == nil {
		return ""
	}

	codecPar := s.CodecPar()
	if s.Type().Is(media.VIDEO) {
		if codecPar.Width() > 0 && codecPar.Height() > 0 {
			return fmt.Sprintf("%dx%d", codecPar.Width(), codecPar.Height())
		}
	}

	if s.Type().Is(media.AUDIO) {
		sampleRate := codecPar.SampleRate()
		channels := codecPar.ChannelLayout().NumChannels()
		switch {
		case sampleRate > 0 && channels > 0:
			return fmt.Sprintf("%dHz, %dch", sampleRate, channels)
		case sampleRate > 0:
			return fmt.Sprintf("%dHz", sampleRate)
		case channels > 0:
			return fmt.Sprintf("%dch", channels)
		}
	}

	if codecPar.BitRate() > 0 {
		return fmt.Sprintf("%dbps", codecPar.BitRate())
	}

	return ""
}
