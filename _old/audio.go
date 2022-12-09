package media

////////////////////////////////////////////////////////////////////////////////
// TYPES

// AudioFormat represents how the samples are stored
type AudioFormat int

// AudioChannelLayout represents number of channels and layout of those channels
type AudioChannelLayout struct {
	Channels uint
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	AUDIO_FMT_NONE AudioFormat = iota
	AUDIO_FMT_U8               // unsigned 8 bits
	AUDIO_FMT_U8P              // unsigned 8 bits, planar
	AUDIO_FMT_S16              // signed 16 bits
	AUDIO_FMT_S16P             // signed 16 bits, planar
	AUDIO_FMT_S32              // signed 32 bits
	AUDIO_FMT_S32P             // signed 32 bits, planar
	AUDIO_FMT_F32              // float32
	AUDIO_FMT_F32P             // float32, planar
	AUDIO_FMT_F64              // float64
	AUDIO_FMT_F64P             // float64, planar
	AUDIO_FMT_S64              // signed 64 bits
	AUDIO_FMT_S64P             // signed 64 bits, planar
)

var (
	AudioLayoutMono   = AudioChannelLayout{1}
	AudioLayoutStereo = AudioChannelLayout{2}
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f AudioFormat) String() string {
	switch f {
	case AUDIO_FMT_NONE:
		return "AUDIO_FMT_NONE"
	case AUDIO_FMT_U8:
		return "AUDIO_FMT_U8"
	case AUDIO_FMT_U8P:
		return "AUDIO_FMT_U8P"
	case AUDIO_FMT_S16:
		return "AUDIO_FMT_S16"
	case AUDIO_FMT_S16P:
		return "AUDIO_FMT_S16P"
	case AUDIO_FMT_S32:
		return "AUDIO_FMT_S32"
	case AUDIO_FMT_S32P:
		return "AUDIO_FMT_S32P"
	case AUDIO_FMT_F32:
		return "AUDIO_FMT_F32"
	case AUDIO_FMT_F32P:
		return "AUDIO_FMT_F32P"
	case AUDIO_FMT_F64:
		return "AUDIO_FMT_F64"
	case AUDIO_FMT_F64P:
		return "AUDIO_FMT_F64P"
	case AUDIO_FMT_S64:
		return "AUDIO_FMT_S64"
	case AUDIO_FMT_S64P:
		return "AUDIO_FMT_S64P"
	default:
		return "[?? Invalid AudioFormat value]"
	}
}
