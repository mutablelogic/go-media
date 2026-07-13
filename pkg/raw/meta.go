package raw

import (
	"encoding/json"
	"fmt"
	"image"
	"strings"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	libraw "github.com/mutablelogic/go-media/sys/libraw"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Meta is a single key/value metadata field from a RAW file.
// Keys follow XMP namespace conventions (tiff:Make, exif:ISOSpeedRatings, etc.).
type Meta struct {
	key string
	val string
	any any
}

var _ media.Metadata = (*Meta)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *Meta) Key() string        { return m.key }
func (m *Meta) Value() string      { return m.val }
func (m *Meta) Bytes() []byte      { return nil }
func (m *Meta) Image() image.Image { return nil }
func (m *Meta) Any() any           { return m.any }

func (m *Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{Key: m.key, Value: m.val})
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE

func newMetadata(r *RAW) []media.Metadata {
	iparams := libraw.Libraw_get_iparams(r.data)
	other := libraw.Libraw_get_imgother(r.data)

	var result []media.Metadata
	add := func(key, val string, typed any) {
		if val != "" {
			result = append(result, &Meta{key: key, val: val, any: typed})
		}
	}

	add("tiff:Make", strings.TrimSpace(libraw.IParams_make(iparams)), strings.TrimSpace(libraw.IParams_make(iparams)))
	add("tiff:Model", strings.TrimSpace(libraw.IParams_model(iparams)), strings.TrimSpace(libraw.IParams_model(iparams)))
	add("tiff:Software", strings.TrimSpace(libraw.IParams_software(iparams)), strings.TrimSpace(libraw.IParams_software(iparams)))

	if iso := libraw.ImgOther_iso_speed(other); iso > 0 {
		add("exif:ISOSpeedRatings", fmt.Sprintf("%.0f", iso), iso)
	}
	if s := libraw.ImgOther_shutter(other); s > 0 {
		add("exif:ExposureTime", shutterStr(s), s)
	}
	if ap := libraw.ImgOther_aperture(other); ap > 0 {
		add("exif:FNumber", fmt.Sprintf("f/%.1f", ap), ap)
	}
	if fl := libraw.ImgOther_focal_len(other); fl > 0 {
		add("exif:FocalLength", fmt.Sprintf("%.0fmm", fl), fl)
	}
	if ts := libraw.ImgOther_timestamp(other); ts > 0 {
		// libraw derives timestamp via mktime() on the EXIF wall-clock
		// string (which carries no timezone), so the epoch is only
		// meaningful when re-localized on the same host that parsed it.
		// Recover the wall-clock fields via time.Local (the inverse of
		// libraw's mktime) and relabel them as UTC, rather than
		// converting, so the result doesn't depend on the host timezone.
		local := time.Unix(ts, 0)
		t := time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), 0, time.UTC)
		add("exif:DateTimeOriginal", t.Format(time.RFC3339), t)
	}
	if desc := strings.TrimSpace(libraw.ImgOther_desc(other)); desc != "" {
		add("dc:description", desc, desc)
	}
	if artist := strings.TrimSpace(libraw.ImgOther_artist(other)); artist != "" {
		add("dc:creator", artist, artist)
	}

	return result
}

func shutterStr(s float32) string {
	if s >= 1 {
		return fmt.Sprintf("%.1fs", s)
	}
	return fmt.Sprintf("1/%.0f", 1.0/s)
}
