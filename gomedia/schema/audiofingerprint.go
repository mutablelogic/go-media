package schema

import (
	"io"

	// Packages
	chromaprintschema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AudioFingerprintRequest struct {
	Input       string    `json:"input,omitempty"` // Input media file path
	Reader      io.Reader `json:"-" kong:"-"`      // Reader for media data
	InputFormat string    `json:"input_format,omitempty" name:"input-format" help:"Input format name (e.g. s16le)"`
	InputOpts   []string  `json:"input_opts,omitempty" name:"input-opt" help:"Input format option key=value (repeatable)"`
	Duration    float64   `json:"duration,omitempty"` // Full track duration in seconds (0 = auto-detect)
}

type AudioFingerprintResponse struct {
	Fingerprint string  `json:"fingerprint"` // Audio fingerprint string
	Duration    float64 `json:"duration"`    // Track duration in seconds
}

type AudioFingerprintLookupRequest struct {
	Fingerprint string   `json:"fingerprint"`        // Audio fingerprint string
	Duration    float64  `json:"duration"`           // Track duration in seconds
	Metadata    []string `json:"metadata,omitempty"` // Metadata to request
}

type AudioFingerprintLookupResponse struct {
	Matches [][]*chromaprintschema.ResponseMatch `json:"matches,omitempty"` // AcoustID matches
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AudioFingerprintResponse) String() string {
	return types.Stringify(r)
}

func (r AudioFingerprintLookupResponse) String() string {
	return types.Stringify(r)
}
