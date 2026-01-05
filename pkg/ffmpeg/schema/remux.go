package schema

import (
	"encoding/json"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type RemuxRequest struct {
	Request
	Output
	Streams      []int             `json:"streams,omitempty"`       // Stream indices to include (empty = all streams)
	CopyMetadata bool              `json:"copy_metadata,omitempty"` // Copy existing metadata from source
	CopyArtwork  bool              `json:"copy_artwork,omitempty"`  // Copy existing artwork from source
	Metadata     map[string]string `json:"metadata,omitempty"`      // Metadata to set (empty string value clears existing)
	Artwork      []Artwork         `json:"artwork,omitempty"`       // Artwork to add/replace
}

type RemuxResponse struct {
	Format   string            `json:"format"`             // Output format name
	Duration float64           `json:"duration"`           // Duration in seconds
	Size     int64             `json:"size"`               // Total bytes written
	Streams  []Stream          `json:"streams,omitempty"`  // Stream information
	Metadata map[string]string `json:"metadata,omitempty"` // Final metadata
	Artwork  []Artwork         `json:"artwork,omitempty"`  // Final artwork
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r RemuxRequest) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r RemuxResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
