package schema

import (
	"encoding/json"

	// Packages
	uuid "github.com/google/uuid"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ProfileMeta struct {
	Description string    `json:"description,omitempty"` // Format description
	Audio       uuid.UUID `json:"audio,omitempty"`       // Audio profile to use for this format
	//	Video    uuid.UUID          `json:"video,omitempty"`    // Video profile to use for this format TODO
	//	Subtitle uuid.UUID          `json:"subtitle,omitempty"` // Subtitle profile to use for this format TODO
	Opts json.RawMessage    `json:"options,omitempty"` // Additional format options
	ctx  *ff.AVOutputFormat `json:"-"`                 // Internal format
	opts map[string]Option  `json:"-"`                 // Internal format options
}

type Profile struct {
	Id   uuid.UUID `json:"id,omitempty"` // Unique identifier for the format profile
	Name string    `json:"name"`         // Format name, e.g. "mp4", "mkv", "flv", ...
	ProfileMeta
}

type ProfileUUID uuid.UUID
