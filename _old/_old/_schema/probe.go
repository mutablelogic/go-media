package schema

import (
	"io"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ProbeRequest struct {
	Reader      io.Reader `json:"-" kong:"-"` // Reader for media data
	InputFormat string    `json:"input_format,omitempty" name:"input-format" help:"Input format name (e.g. mpegts)"`
	InputOpts   []string  `json:"input_opts,omitempty" name:"input-opt" help:"Input format option key=value (repeatable)"`
}

type ProbeResponse struct {
	Format      string    `json:"format"`                // Format name (e.g., "mov,mp4,m4a,3gp,3g2,mj2")
	Description string    `json:"description,omitempty"` // Format description (e.g., "QuickTime / MOV")
	MimeTypes   []string  `json:"mime_types,omitempty"`  // MIME types
	Duration    float64   `json:"duration"`              // Duration in seconds
	Streams     []*Stream `json:"streams,omitempty"`     // Stream information
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r ProbeResponse) String() string {
	return types.Stringify(r)
}
