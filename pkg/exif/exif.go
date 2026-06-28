package exif

import (
	"encoding/json"
	"io"
	"os"

	// Packages
	media "github.com/mutablelogic/go-media"
	libexif "github.com/mutablelogic/go-media/sys/libexif"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type EXIF struct {
	data  *libexif.Data
	order libexif.ByteOrder
}

var _ io.Closer = (*EXIF)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func Open(path string) (*EXIF, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, media.ErrNotFound.Withf("%q", path)
	}
	data := libexif.Exif_data_new_from_file(path)
	if data == nil {
		return nil, media.ErrBadParameter.Withf("no EXIF data in %q", path)
	}
	return newEXIF(data), nil
}

func Read(r io.Reader) (*EXIF, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(buf)
}

func Parse(data []byte) (*EXIF, error) {
	if len(data) == 0 {
		return nil, media.ErrBadParameter.With("empty data")
	}
	d := libexif.Exif_data_new_from_data(data)
	if d == nil {
		return nil, media.ErrBadParameter.With("failed to parse EXIF data")
	}
	return newEXIF(d), nil
}

func (e *EXIF) Close() error {
	if e.data != nil {
		libexif.Exif_data_unref(e.data)
		e.data = nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// MarshalJSON implements json.Marshaler.
func (e *EXIF) MarshalJSON() ([]byte, error) {
	v := struct {
		Order      string `json:"order"`
		Tags       []*Tag `json:"tags"`
		MakerNote  *MakerNote `json:"maker_note,omitempty"`
	}{
		Order:     libexif.Exif_byte_order_get_name(e.order),
		Tags:      e.Tags(),
		MakerNote: e.MakerNote(),
	}
	return json.Marshal(v)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e *EXIF) MakerNote() *MakerNote {
	mn := libexif.Exif_data_get_maker_note_data(e.data)
	if mn == nil {
		return nil
	}
	return newMakerNote(mn)
}

func (e *EXIF) Tags() []*Tag {
	var tags []*Tag
	for ifd := libexif.EXIF_IFD_0; ifd < libexif.EXIF_IFD_COUNT; ifd++ {
		content := libexif.Exif_data_get_content(e.data, ifd)
		if content == nil {
			continue
		}
		libexif.Exif_content_foreach_entry(content, func(entry *libexif.Entry) {
			tags = append(tags, newTag(entry, ifd, e.order))
		})
	}
	return tags
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newEXIF(data *libexif.Data) *EXIF {
	return &EXIF{
		data:  data,
		order: libexif.Exif_data_get_byte_order(data),
	}
}
