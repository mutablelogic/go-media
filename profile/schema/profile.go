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
