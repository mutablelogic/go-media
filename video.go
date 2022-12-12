package media

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PixelFormat specifies the encoding of pixel within a frame
type PixelFormat int

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	PIXEL_FORMAT_YUV420P PixelFormat = iota
	PIXEL_FORMAT_NONE    PixelFormat = -1
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f PixelFormat) String() string {
	switch f {
	case PIXEL_FORMAT_NONE:
		return "PIXEL_FORMAT_NONE"
	case PIXEL_FORMAT_YUV420P:
		return "PIXEL_FORMAT_YUV420P"
	default:
		return "[?? Invalid PixelFormat value]"
	}
}
