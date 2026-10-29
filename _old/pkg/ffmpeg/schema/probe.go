package schema

import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ProbeRequest struct {
	Request
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
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
