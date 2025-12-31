package schema

import (
	"encoding/json"

	// Packages
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AudioFingerprintRequest struct {
	Request
	Duration float64  `json:"duration,omitempty"` // Full track duration in seconds (0 = auto-detect)
	Lookup   bool     `json:"lookup,omitempty"`   // Perform AcoustID lookup
	Metadata []string `json:"metadata,omitempty"` // Metadata to request: "recordings", "recordingids", "releases", "releaseids", "releasegroups", "releasegroupids", "tracks", "compress", "usermeta", "sources"
}

type AudioFingerprintResponse struct {
	Fingerprint string                      `json:"fingerprint"`       // Audio fingerprint string
	Duration    float64                     `json:"duration"`          // Track duration in seconds
	Matches     []chromaprint.ResponseMatch `json:"matches,omitempty"` // AcoustID matches (if lookup=true)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AudioFingerprintRequest) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r AudioFingerprintResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
