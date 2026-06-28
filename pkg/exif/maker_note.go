package exif

import (
	"encoding/json"

	libexif "github.com/mutablelogic/go-media/sys/libexif"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MakerNote struct {
	d *libexif.MakerNoteData
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE CONSTRUCTOR

func newMakerNote(d *libexif.MakerNoteData) *MakerNote {
	return &MakerNote{d: d}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// MarshalJSON implements json.Marshaler.
func (m *MakerNote) MarshalJSON() ([]byte, error) {
	count := libexif.Exif_mnote_data_count(m.d)
	entries := make([]struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
		Value       string `json:"value"`
	}, count)
	for i := range entries {
		n := uint(i)
		entries[i].ID = libexif.Exif_mnote_data_get_id(m.d, n)
		entries[i].Name = libexif.Exif_mnote_data_get_name(m.d, n)
		entries[i].Title = libexif.Exif_mnote_data_get_title(m.d, n)
		entries[i].Description = libexif.Exif_mnote_data_get_description(m.d, n)
		entries[i].Value = libexif.Exif_mnote_data_get_value(m.d, n)
	}
	return json.Marshal(entries)
}
