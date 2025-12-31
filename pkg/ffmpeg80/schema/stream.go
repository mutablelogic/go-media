package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Stream represents an AVStream from a media file
type Stream struct {
	Index       int      `json:"index"`                 // Stream index in the container
	ID          int      `json:"id,omitempty"`          // Stream ID (format-specific)
	Type        string   `json:"type"`                  // "video", "audio", "subtitle", "data", "attachment"
	Codec       string   `json:"codec,omitempty"`       // Codec name
	CodecID     string   `json:"codec_id,omitempty"`    // Codec ID
	CodecTag    string   `json:"codec_tag,omitempty"`   // Codec tag (fourcc)
	BitRate     int64    `json:"bit_rate,omitempty"`    // Bit rate in bits/s
	Duration    float64  `json:"duration,omitempty"`    // Duration in seconds
	StartTime   float64  `json:"start_time,omitempty"`  // Start time in seconds
	NumFrames   int64    `json:"num_frames,omitempty"`  // Number of frames (if known)
	TimeBase    string   `json:"time_base,omitempty"`   // Time base as "num/den"
	Disposition []string `json:"disposition,omitempty"` // Disposition flags

	// Video-specific fields
	Width             int    `json:"width,omitempty"`               // Video width
	Height            int    `json:"height,omitempty"`              // Video height
	PixelFormat       string `json:"pixel_format,omitempty"`        // Pixel format name
	SampleAspectRatio string `json:"sample_aspect_ratio,omitempty"` // Sample aspect ratio as "num:den"

	// Audio-specific fields
	SampleRate    int    `json:"sample_rate,omitempty"`    // Audio sample rate in Hz
	SampleFormat  string `json:"sample_format,omitempty"`  // Sample format name
	Channels      int    `json:"channels,omitempty"`       // Number of audio channels
	ChannelLayout string `json:"channel_layout,omitempty"` // Channel layout name
	FrameSize     int    `json:"frame_size,omitempty"`     // Audio frame size
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewStream creates a Stream from an AVStream
func NewStream(stream *ff.AVStream) *Stream {
	if stream == nil {
		return nil
	}

	codecPar := stream.CodecPar()
	if codecPar == nil {
		return nil
	}

	// Get codec type
	codecType := codecPar.CodecType()
	typeStr := mediaTypeString(codecType)

	// Get codec name from ID
	codecID := codecPar.CodecID()
	var codecName string
	if codec := ff.AVCodec_find_decoder(codecID); codec != nil {
		codecName = codec.Name()
	}

	// Get codec tag as fourcc string
	codecTag := codecPar.CodecTag()
	var codecTagStr string
	if codecTag != 0 {
		codecTagStr = codecTagToString(codecTag)
	}

	// Get time base
	timeBase := stream.TimeBase()
	var timeBaseStr string
	if timeBase.Den() != 0 {
		timeBaseStr = rationalToString(timeBase)
	}

	// Calculate duration and start time in seconds
	var duration, startTime float64
	if timeBase.Den() != 0 {
		// Get raw timestamps from the stream
		// AVStream stores start_time and duration as int64 in time_base units
		// We need to convert them to seconds using the time base
		streamDuration := float64(stream.Duration()) * float64(timeBase.Num()) / float64(timeBase.Den())
		streamStartTime := float64(stream.StartTime()) * float64(timeBase.Num()) / float64(timeBase.Den())
		if streamDuration > 0 {
			duration = streamDuration
		}
		if streamStartTime >= 0 && streamStartTime < 1e9 { // Sanity check
			startTime = streamStartTime
		}
	}

	// Get disposition flags
	disposition := stream.Disposition()
	var dispositionFlags []string
	if dispStr := disposition.String(); dispStr != "" {
		dispositionFlags = strings.Split(dispStr, "|")
	}

	s := &Stream{
		Index:       stream.Index(),
		ID:          stream.Id(),
		Type:        typeStr,
		Codec:       codecName,
		CodecID:     codecID.Name(),
		CodecTag:    codecTagStr,
		BitRate:     codecPar.BitRate(),
		Duration:    duration,
		StartTime:   startTime,
		NumFrames:   stream.NumFrames(),
		TimeBase:    timeBaseStr,
		Disposition: dispositionFlags,
	}

	// Add type-specific fields
	switch codecType {
	case ff.AVMEDIA_TYPE_VIDEO:
		s.Width = codecPar.Width()
		s.Height = codecPar.Height()
		if pf := codecPar.PixelFormat(); pf != ff.AV_PIX_FMT_NONE {
			s.PixelFormat = ff.AVUtil_get_pix_fmt_name(pf)
		}
		if sar := codecPar.SampleAspectRatio(); sar.Num() != 0 || sar.Den() != 0 {
			s.SampleAspectRatio = rationalToString(sar)
		}
	case ff.AVMEDIA_TYPE_AUDIO:
		s.SampleRate = codecPar.SampleRate()
		if sf := codecPar.SampleFormat(); sf != ff.AV_SAMPLE_FMT_NONE {
			s.SampleFormat = ff.AVUtil_get_sample_fmt_name(sf)
		}
		layout := codecPar.ChannelLayout()
		s.Channels = layout.NumChannels()
		if name, err := ff.AVUtil_channel_layout_describe(&layout); err == nil && name != "" {
			s.ChannelLayout = name
		}
		s.FrameSize = codecPar.FrameSize()
	}

	return s
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

// IsDefault returns true if this stream has the default disposition
func (s *Stream) IsDefault() bool {
	for _, d := range s.Disposition {
		if d == "DEFAULT" {
			return true
		}
	}
	return false
}

// IsAttachedPic returns true if this stream is an attached picture (album art)
func (s *Stream) IsAttachedPic() bool {
	for _, d := range s.Disposition {
		if d == "ATTACHED_PIC" {
			return true
		}
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s Stream) String() string {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE HELPERS

// rationalToString formats an AVRational as "num/den"
func rationalToString(r ff.AVRational) string {
	return fmt.Sprintf("%d/%d", r.Num(), r.Den())
}

// codecTagToString converts a codec tag (fourcc) to a readable string
func codecTagToString(tag uint32) string {
	b := make([]byte, 4)
	b[0] = byte(tag)
	b[1] = byte(tag >> 8)
	b[2] = byte(tag >> 16)
	b[3] = byte(tag >> 24)
	// Only return if all bytes are printable ASCII
	for _, c := range b {
		if c < 0x20 || c > 0x7e {
			return ""
		}
	}
	return string(b)
}
