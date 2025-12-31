package schema

import (
	"io"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Request struct {
	Path   string    `json:"path"` // Path to media file
	Reader io.Reader `json:"-"`    // Reader for media data
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *Request) Open(opt ...ffmpeg.Opt) (*ffmpeg.Reader, error) {
	// Open by reader if set
	if r.Reader != nil {
		if r.Path != "" {
			opt = append(opt, ffmpeg.OptInputFormat(r.Path))
		}
		return ffmpeg.NewReader(r.Reader, opt...)
	}
	// Open by path otherwise
	return ffmpeg.Open(r.Path, opt...)
}
