package media

import ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"

////////////////////////////////////////////////////////////////////////////
// TYPES

type manager struct {
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager() *manager {
	return new(manager)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the list of input formats, filtering by name or mimetype
func (this *manager) InputFormats(mimetype string) []InputFormat {
	var result []InputFormat
	var opaque uintptr
	for {
		demuxer := ffmpeg.AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		if this.matchesInput(demuxer, mimetype) {
			result = append(result, demuxer)
		}
	}
	return result

}

// Return the list of output formats, filtering by name or mimetype
func (this *manager) OutputFormats(name string) []OutputFormat {
	return nil
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *manager) matchesInput(demuxer *ffmpeg.AVInputFormat, mimetype string) bool {
	return true
}
