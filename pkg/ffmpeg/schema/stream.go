package schema

import (
	"encoding/json"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Stream represents a thin wrapper around AVStream
type Stream struct {
	*ff.AVStream
	codecPar ff.AVCodecParameters // Copied at construction to survive reader close
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewStream creates a Stream from an AVStream
func NewStream(stream *ff.AVStream) *Stream {
	if stream == nil {
		return nil
	}
	s := &Stream{AVStream: stream}
	// Copy codec parameters so they remain valid after reader is closed
	if codecPar := stream.CodecPar(); codecPar != nil {
		s.codecPar = *codecPar
	}
	return s
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Stream) MarshalJSON() ([]byte, error) {
	if s.AVStream == nil {
		return json.Marshal(nil)
	}

	// Build a proper JSON representation with meaningful fields
	type streamJSON struct {
		Index      int     `json:"index"`
		ID         int     `json:"id"`
		Type       string  `json:"type"`
		CodecName  string  `json:"codec_name,omitempty"`
		CodecType  string  `json:"codec_type,omitempty"`
		Width      int     `json:"width,omitempty"`
		Height     int     `json:"height,omitempty"`
		SampleRate int     `json:"sample_rate,omitempty"`
		Channels   int     `json:"channels,omitempty"`
		StartTime  float64 `json:"start_time"`
		Duration   float64 `json:"duration"`
		BitRate    int64   `json:"bit_rate,omitempty"`
	}

	obj := streamJSON{
		Index: s.AVStream.Index(),
		ID:    s.AVStream.Id(),
		Type:  s.Type().String(),
	}

	// Add codec info
	if codecPar := s.CodecPar(); codecPar != nil {
		obj.CodecType = codecPar.CodecType().String()
		if codec := ff.AVCodec_find_decoder(codecPar.CodecID()); codec != nil {
			obj.CodecName = codec.Name()
		}
		obj.BitRate = codecPar.BitRate()

		// Type-specific fields
		switch codecPar.CodecType() {
		case ff.AVMEDIA_TYPE_VIDEO:
			obj.Width = codecPar.Width()
			obj.Height = codecPar.Height()
		case ff.AVMEDIA_TYPE_AUDIO:
			obj.SampleRate = codecPar.SampleRate()
			obj.Channels = codecPar.ChannelLayout().NumChannels()
		}
	}

	// Time information
	tb := s.AVStream.TimeBase()
	if tb.Num() > 0 && tb.Den() > 0 {
		if start := s.AVStream.StartTime(); start != int64(ff.AV_NOPTS_VALUE) {
			obj.StartTime = float64(start) * ff.AVUtil_rational_q2d(tb)
		}
		if duration := s.AVStream.Duration(); duration != int64(ff.AV_NOPTS_VALUE) && duration > 0 {
			obj.Duration = float64(duration) * ff.AVUtil_rational_q2d(tb)
		}
	}

	return json.Marshal(obj)
}

func (s *Stream) String() string {
	if s == nil || s.AVStream == nil {
		return "<nil>"
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Type returns the media type of this stream
func (s *Stream) Type() media.Type {
	if s == nil {
		return media.UNKNOWN
	}
	// Check for attached picture (album art)
	if s.AVStream != nil && s.AVStream.Disposition()&ff.AV_DISPOSITION_ATTACHED_PIC != 0 {
		return media.DATA
	}
	switch s.codecPar.CodecType() {
	case ff.AVMEDIA_TYPE_VIDEO:
		return media.VIDEO
	case ff.AVMEDIA_TYPE_AUDIO:
		return media.AUDIO
	case ff.AVMEDIA_TYPE_DATA:
		return media.DATA
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return media.SUBTITLE
	default:
		return media.UNKNOWN
	}
}

// CodecPar returns the codec parameters for this stream
func (s *Stream) CodecPar() *ff.AVCodecParameters {
	if s == nil {
		return nil
	}
	return &s.codecPar
}
