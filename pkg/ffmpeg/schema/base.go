package schema

import (
	"io"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Request struct {
	Input  string    `json:"input" arg:""` // Input media file path
	Reader io.Reader `json:"-" kong:"-"`   // Reader for media data
}

type Output struct {
	Output     string   `json:"output"  arg:""` // Output media file path
	OutputOpts []string `json:"opts,omitempty"` // Format-specific options (e.g., "movflags=+faststart")
}

// Writer is an optional interface that writers can implement
// to receive progress updates during long-running operations
type Writer interface {
	io.Writer
	Progress(current, total int64) // Report progress (units depends on task)
	Log(message string)            // Receive log messages
}

// FrameWriter is an optional interface that writers can implement
// to receive decoded frames directly instead of JSON output
type FrameWriter interface {
	io.Writer
	WriteFrame(streamIndex int, frame interface{}) error
}
