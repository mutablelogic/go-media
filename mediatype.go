package media

////////////////////////////////////////////////////////////////////////////
// TYPES

// Media type flags
type MediaType uint32

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	UNKNOWN  MediaType = (1 << iota) // Usually treated as DATA
	VIDEO                            // Video stream
	AUDIO                            // Audio stream
	DATA                             // Opaque data information usually continuous
	SUBTITLE                         // Subtitle stream
	INPUT                            // Demuxer
	OUTPUT                           // Muxer
	DEVICE                           // Device rather than byte stream or file
)
