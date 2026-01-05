package schema

import (
	"encoding/json"
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
	Fingerprint string                  `json:"fingerprint"`       // Audio fingerprint string
	Duration    float64                 `json:"duration"`          // Track duration in seconds
	Matches     []AudioFingerprintMatch `json:"matches,omitempty"` // AcoustID matches (if lookup=true)
}

type AudioFingerprintMatch struct {
	Id         string                      `json:"id"`
	Score      float64                     `json:"score"`
	Recordings []AudioFingerprintRecording `json:"recordings,omitempty"`
}

type AudioFingerprintRecording struct {
	Id            string                   `json:"id"`
	Title         string                   `json:"title,omitempty"`
	Duration      float64                  `json:"duration,omitempty"`
	Artists       []AudioFingerprintArtist `json:"artists,omitempty"`
	ReleaseGroups []AudioFingerprintGroup  `json:"releasegroups,omitempty"`
}

type AudioFingerprintArtist struct {
	Id   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type AudioFingerprintGroup struct {
	Id       string                    `json:"id"`
	Type     string                    `json:"type,omitempty"`
	Title    string                    `json:"title,omitempty"`
	Releases []AudioFingerprintRelease `json:"releases,omitempty"`
}

type AudioFingerprintRelease struct {
	Id      string                   `json:"id"`
	Mediums []AudioFingerprintMedium `json:"mediums,omitempty"`
}

type AudioFingerprintMedium struct {
	Position float64                 `json:"position,omitempty"`
	Tracks   []AudioFingerprintTrack `json:"tracks,omitempty"`
}

type AudioFingerprintTrack struct {
	Position float64 `json:"position,omitempty"`
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
