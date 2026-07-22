package schema

import (
	"encoding/json"
	"strings"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ProfileMeta struct {
	Description string             `json:"description,omitempty"` // Format description
	Opts        json.RawMessage    `json:"options,omitempty"`     // Additional format options
	ctx         *ff.AVOutputFormat `json:"-"`                     // Internal format
	opts        map[string]Option  `json:"-"`                     // Internal format options
}

type Profile struct {
	Id     uuid.UUID `json:"id,omitempty"` // Unique identifier for the format profile
	Format string    `json:"format"`       // Format name, e.g. "mp4", "mkv", "flv", ...
	ProfileMeta
}

type ProfileUUID uuid.UUID

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProfile(name string, req ProfileMeta) (*Profile, error) {
	self := new(Profile)

	// Get AVOutputFormat for the given format name
	format := ff.AVFormat_guess_format(name, "", "")
	if format == nil {
		return nil, gomedia.ErrNotFound.Withf("format %q", name)
	} else {
		self.ctx = format
	}

	// Set description
	if desc := strings.TrimSpace(req.Description); desc != "" {
		self.Description = desc
	} else {
		self.Description = format.LongName()
	}

	// Return success
	return self, nil
}
