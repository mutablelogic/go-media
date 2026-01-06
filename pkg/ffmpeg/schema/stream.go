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

	// Build JSON with the saved codec parameters
	type jsonStream struct {
		Index       int                   `json:"index"`
		Id          int                   `json:"id"`
		CodecPar    *ff.AVCodecParameters `json:"codec_par,omitempty"`
		StartTime   ff.AVTimestamp        `json:"start_time"`
		Duration    ff.AVTimestamp        `json:"duration"`
		NumFrames   int64                 `json:"num_frames,omitempty"`
		TimeBase    ff.AVRational         `json:"time_base,omitempty"`
		Disposition ff.AVDisposition      `json:"disposition,omitempty"`
	}

	return json.Marshal(jsonStream{
		Index:       s.AVStream.Index(),
		Id:          s.AVStream.Id(),
		CodecPar:    &s.codecPar, // Use saved copy
		StartTime:   ff.AVTimestamp(s.AVStream.StartTime()),
		Duration:    ff.AVTimestamp(s.AVStream.Duration()),
		NumFrames:   s.AVStream.NumFrames(),
		TimeBase:    s.AVStream.TimeBase(),
		Disposition: s.AVStream.Disposition(),
	})
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
