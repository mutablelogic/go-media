package schema

import (
	"encoding/json"

	uuid "github.com/google/uuid"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Profile interface {
	UUID() uuid.UUID            // Unique identifier for this profile
	Type() CodecType            // Type for this profile
	Codec() *Codec              // Codec for this profile
	Par() *ff.AVCodecParameters // Codec parameters for this profile
	TimeBase() *ff.AVRational   // Time base for this profile
	Options() json.RawMessage   // Additional codec-specific options
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	OptionBitrate = "bitrate"

	// Audio options
	OptionProfile       = "profile"
	OptionSampleRate    = "sample_rate"
	OptionSampleFormat  = "sample_format"
	OptionChannelLayout = "channel_layout"

	// Video options
	OptionWidth       = "width"
	OptionHeight      = "height"
	OptionPixelFormat = "pixel_format"
	OptionFrameRate   = "frame_rate"
)
