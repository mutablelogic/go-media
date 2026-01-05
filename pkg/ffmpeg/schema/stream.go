package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Stream represents an AVStream from a media file
type Stream struct {
	ctx *ff.AVStream
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new stream
func newStream(ctx *ff.AVStream) *Stream {
	return &Stream{ctx}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Stream) MarshalJSON() ([]byte, error) {
	// Stream represents an AVStream from a media file
	type jsonStream struct {
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

	codecPar := s.ctx.CodecPar()
	if codecPar == nil {
		return json.Marshal(jsonStream{
			Index: s.Index(),
		})
	}

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
	timeBase := s.ctx.TimeBase()
	var timeBaseStr string
	if timeBase.Den() != 0 {
		timeBaseStr = rationalToString(timeBase)
	}

	// Calculate duration and start time in seconds
	var duration, startTime float64
	if timeBase.Den() != 0 {
		streamDuration := float64(s.ctx.Duration()) * float64(timeBase.Num()) / float64(timeBase.Den())
		streamStartTime := float64(s.ctx.StartTime()) * float64(timeBase.Num()) / float64(timeBase.Den())
		if streamDuration > 0 {
			duration = streamDuration
		}
		if streamStartTime >= 0 && streamStartTime < 1e9 {
			startTime = streamStartTime
		}
	}

	// Get disposition flags
	disposition := s.ctx.Disposition()
	var dispositionFlags []string
	if dispStr := disposition.String(); dispStr != "" {
		dispositionFlags = strings.Split(dispStr, "|")
	}

	result := jsonStream{
		Index:       s.Index(),
		ID:          s.ctx.Id(),
		Type:        typeStr,
		Codec:       codecName,
		CodecID:     codecID.Name(),
		CodecTag:    codecTagStr,
		BitRate:     codecPar.BitRate(),
		Duration:    duration,
		StartTime:   startTime,
		NumFrames:   s.ctx.NumFrames(),
		TimeBase:    timeBaseStr,
		Disposition: dispositionFlags,
	}

	// Add type-specific fields
	switch codecType {
	case ff.AVMEDIA_TYPE_VIDEO:
		result.Width = codecPar.Width()
		result.Height = codecPar.Height()
		if pf := codecPar.PixelFormat(); pf != ff.AV_PIX_FMT_NONE {
			result.PixelFormat = ff.AVUtil_get_pix_fmt_name(pf)
		}
		if sar := codecPar.SampleAspectRatio(); sar.Num() != 0 || sar.Den() != 0 {
			result.SampleAspectRatio = rationalToString(sar)
		}
	case ff.AVMEDIA_TYPE_AUDIO:
		result.SampleRate = codecPar.SampleRate()
		if sf := codecPar.SampleFormat(); sf != ff.AV_SAMPLE_FMT_NONE {
			result.SampleFormat = ff.AVUtil_get_sample_fmt_name(sf)
		}
		layout := codecPar.ChannelLayout()
		result.Channels = layout.NumChannels()
		if name, err := ff.AVUtil_channel_layout_describe(&layout); err == nil && name != "" {
			result.ChannelLayout = name
		}
		result.FrameSize = codecPar.FrameSize()
	}

	return json.Marshal(result)
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Return the stream index
func (s *Stream) Index() int {
	return int(s.ctx.Index())
}

// Return the stream type
func (s *Stream) Type() media.Type {
	if s.ctx.Disposition()&ff.AV_DISPOSITION_ATTACHED_PIC != 0 {
		return media.DATA
	}
	codecPar := s.ctx.CodecPar()
	if codecPar == nil {
		return media.UNKNOWN
	}
	switch codecPar.CodecType() {
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

// Return the codec parameters
func (s *Stream) CodecPar() *ff.AVCodecParameters {
	return s.ctx.CodecPar()
}

// NewStream creates a Stream from an AVStream
func NewStream(stream *ff.AVStream) *Stream {
	return newStream(stream)
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

// IsDefault returns true if this stream has the default disposition
func (s *Stream) IsDefault() bool {
	return s.ctx.Disposition()&ff.AV_DISPOSITION_DEFAULT != 0
}

// IsAttachedPic returns true if this stream is an attached picture (album art)
func (s *Stream) IsAttachedPic() bool {
	return s.ctx.Disposition()&ff.AV_DISPOSITION_ATTACHED_PIC != 0
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
