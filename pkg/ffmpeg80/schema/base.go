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

type Output struct {
	Format     string   `json:"format"`         // Output container format (e.g., "mp4", "mkv", "mov")
	FormatOpts []string `json:"opts,omitempty"` // Format-specific options (e.g., "movflags=+faststart")
}

// Writer is an optional interface that writers can implement
// to receive progress updates during long-running operations
type Writer interface {
	io.Writer
	Progress(current, total int64) // Report progress (units depends on task)
	Log(message string)            // Receive log messages
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *Request) Open(opt ...ffmpeg.Opt) (*ffmpeg.Reader, error) {
	// Open by reader if set
	if r.Reader != nil {
		if r.Path != "" {
			opt = append(opt, ffmpeg.WithInput(r.Path))
		}
		return ffmpeg.NewReader(r.Reader, opt...)
	}
	// Open by path otherwise
	return ffmpeg.Open(r.Path, opt...)
}
