package schema

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Artwork represents embedded artwork (cover art, thumbnails, etc.) as raw bytes.
// When JSON encoded, it is represented as a base64 string.
type Artwork []byte

type ProbeRequest struct {
	Request
	Metadata bool `json:"metadata,omitempty"` // Include metadata in response
	Artwork  bool `json:"artwork,omitempty"`  // Include artwork in response (base64 encoded)
}

type ProbeResponse struct {
	Format      string            `json:"format"`                // Format name (e.g., "mov,mp4,m4a,3gp,3g2,mj2")
	Description string            `json:"description,omitempty"` // Format description (e.g., "QuickTime / MOV")
	MimeTypes   []string          `json:"mime_types,omitempty"`  // MIME types
	Duration    float64           `json:"duration"`              // Duration in seconds
	Streams     []Stream          `json:"streams,omitempty"`     // Stream information
	Metadata    map[string]string `json:"metadata,omitempty"`    // Metadata key-value pairs
	Artwork     []Artwork         `json:"artwork,omitempty"`     // Artwork images as raw bytes (JSON encodes as base64)
}
