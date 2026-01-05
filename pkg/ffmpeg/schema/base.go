package schema

import (
	"io"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Request struct {
	Path   string    `json:"path" arg:""` // Path to media file
	Reader io.Reader `json:"-" kong:"-"`  // Reader for media data
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
