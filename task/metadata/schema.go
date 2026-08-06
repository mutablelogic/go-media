package metadata

import (
	"io"
	"net/url"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
	task "github.com/mutablelogic/go-media/task/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataRequest struct {
	Reader io.Reader `json:"-"`
}

type MetadataResponse struct {
	Name     string             `json:"name,omitempty" help:"Name of the probed input, if known (e.g. the uploaded filename)." example:"sample.mp3"`
	Type     string             `json:"type,omitempty" help:"Detected content type (MIME type) of the input, e.g. \"video/mp4\"." example:"video/mp4"`
	Metadata []profile.Metadata `json:"metadata,omitempty" help:"Container-level metadata tags, e.g. \"title\", \"artist\"; excludes artwork and chapters."`
	Artwork  []profile.Artwork  `json:"artwork,omitempty" help:"Artwork images embedded in the input, e.g. album cover art."`
}

var _ task.Task = (*MetadataRequest)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

// Query returns no parameters - Reader is excluded from JSON (it's supplied
// as the request body, not the query string), leaving nothing to encode.
func (r MetadataRequest) Query() url.Values {
	return url.Values{}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - TASK

// Task identifies this task's kind, for Status.Task.
func (r *MetadataRequest) Task() string {
	return "metadata"
}

// Validate checks that r is well-formed: a non-nil Reader.
func (r *MetadataRequest) Validate() error {
	if r.Reader == nil {
		return gomedia.ErrBadParameter.With("nil reader")
	}
	return nil
}
