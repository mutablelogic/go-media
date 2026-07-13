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
	if stripped := unwrapHEIFExif(data); len(stripped) > 0 {
		if d := parseData(stripped); d != nil {
			return newEXIF(d), nil
		}
	}
	if d := parseData(data); d != nil {
		return newEXIF(d), nil
	}
	if len(data) > 4 {
		if d := parseData(data[4:]); d != nil {
			return newEXIF(d), nil
		}
	}
	return nil, media.ErrBadParameter.With("failed to parse EXIF data")
}

func unwrapHEIFExif(data []byte) []byte {
	if len(data) >= 4 {
		return data[4:]
	}
	return nil
}

func parseData(data []byte) *libexif.Data {
	loader := libexif.Exif_loader_new()
	if loader == nil {
		if d := libexif.Exif_data_new_from_data(data); d != nil {
			return d
		}
		if d := libexif.Exif_data_new(); d != nil {
			libexif.Exif_data_load_data(d, data)
			return d
		}
		return nil
	}
	defer libexif.Exif_loader_unref(loader)

	libexif.Exif_loader_write(loader, data)
	if d := libexif.Exif_loader_get_data(loader); d != nil {
		return d
	}

	if d := libexif.Exif_data_new_from_data(data); d != nil {
		return d
	}

	if d := libexif.Exif_data_new(); d != nil {
		libexif.Exif_data_load_data(d, data)
		return d
	}

	return nil
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
		Order     string     `json:"order"`
		Tags      []*Tag     `json:"tags"`
		MakerNote *MakerNote `json:"maker_note,omitempty"`
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
