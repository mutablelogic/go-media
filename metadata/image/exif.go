package image

import (
	"context"
	"image"
	"io"
	"maps"
	"regexp"
	"strings"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	exif "github.com/mutablelogic/go-media/pkg/exif"
	libexif "github.com/mutablelogic/go-media/sys/libexif"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// exifDateTimeTag associates an EXIF date/time tag with its corresponding
// UTC offset tag, keyed by their "namespace:name" tag key. The offset key
// is empty if the tag has no corresponding offset tag.
type exifDateTimeTag struct {
	dateKey   string
	offsetKey string
}

// timeMetadata wraps a parsed time.Time as gomedia.Metadata, replacing the
// raw EXIF date/time string tag it was parsed from.
type timeMetadata struct {
	key string
	t   time.Time
}

// exifGPSTag associates an EXIF GPS DMS (degrees/minutes/seconds) tag with
// its hemisphere reference tag, keyed by their "namespace:name" tag key.
// negRef is the reference value ("S" or "W") which makes the resulting
// decimal degrees negative.
type exifGPSTag struct {
	dmsKey string
	refKey string
	negRef string
}

// floatMetadata wraps a parsed float64 as gomedia.Metadata, replacing the
// raw EXIF rational tag it was parsed from. Value() returns the original
// libexif-formatted string (e.g. "f/2.8", "40, 44, 54.36999999999999"),
// while Any() returns the parsed float64.
type floatMetadata struct {
	key   string
	value float64
	str   string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var exifDateTimeTags = []exifDateTimeTag{
	{"tiff:DateTime", "exif:OffsetTime"},
	{"exif:DateTimeOriginal", "exif:OffsetTimeOriginal"},
	{"exif:DateTimeDigitized", "exif:OffsetTimeDigitized"},
}

var exifGPSTags = []exifGPSTag{
	{"exif:GPSLatitude", "exif:GPSLatitudeRef", "S"},
	{"exif:GPSLongitude", "exif:GPSLongitudeRef", "W"},
}

////////////////////////////////////////////////////////////////////////////////
// METADATA INTERFACE

func (v timeMetadata) Key() string        { return v.key }
func (v timeMetadata) Value() string      { return v.t.Format(time.RFC3339) }
func (v timeMetadata) Bytes() []byte      { return nil }
func (v timeMetadata) Image() image.Image { return nil }
func (v timeMetadata) Any() any           { return v.t }

func (v floatMetadata) Key() string        { return v.key }
func (v floatMetadata) Value() string      { return v.str }
func (v floatMetadata) Bytes() []byte      { return nil }
func (v floatMetadata) Image() image.Image { return nil }
func (v floatMetadata) Any() any           { return v.value }

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// parseExifDateTime parses an EXIF date/time string (format
// "2006:01:02 15:04:05"), optionally appending a UTC offset (format
// "-07:00") if present.
func parseExifDateTime(dt, offset string) (time.Time, bool) {
	layout := "2006:01:02 15:04:05"
	if offset != "" {
		layout += "-07:00"
		dt += offset
	}
	t, err := time.Parse(layout, dt)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// dmsToDecimal converts a 3-component EXIF DMS (degrees, minutes, seconds)
// rational value into decimal degrees.
func dmsToDecimal(dms []libexif.Rational) (float64, bool) {
	if len(dms) != 3 {
		return 0, false
	}
	deg, ok := rationalToFloat(dms[0])
	if !ok {
		return 0, false
	}
	min, ok := rationalToFloat(dms[1])
	if !ok {
		return 0, false
	}
	sec, ok := rationalToFloat(dms[2])
	if !ok {
		return 0, false
	}
	return deg + min/60 + sec/3600, true
}

// rationalToFloat converts an EXIF rational value to a float64.
func rationalToFloat(r libexif.Rational) (float64, bool) {
	if r.Denominator == 0 {
		return 0, false
	}
	return float64(r.Numerator) / float64(r.Denominator), true
}

// srationalToFloat converts a signed EXIF rational value to a float64.
func srationalToFloat(r libexif.SRational) (float64, bool) {
	if r.Denominator == 0 {
		return 0, false
	}
	return float64(r.Numerator) / float64(r.Denominator), true
}

// singleRationalToFloat converts a tag whose value is a single
// RATIONAL or SRATIONAL component to a float64.
func singleRationalToFloat(tag *exif.Tag) (float64, bool) {
	switch v := tag.Any().(type) {
	case libexif.Rational:
		return rationalToFloat(v)
	case libexif.SRational:
		return srationalToFloat(v)
	default:
		return 0, false
	}
}

////////////////////////////////////////////////////////////////////////////////
// SHARED HELPERS

// exifTagsToMetadata converts raw EXIF tags into gomedia.Metadata, keyed by
// "namespace:name" tag key, replacing date/time, GPS and other rational
// tags with their parsed equivalents (see timeMetadata, floatMetadata).
// It is shared by the JPEG handler and, for RAW files, the embedded
// thumbnail's EXIF data.
func exifTagsToMetadata(tags []*exif.Tag) map[string]gomedia.Metadata {
	// Create a map of tags first
	tagmap := maps.Collect(func(yield func(string, *exif.Tag) bool) {
		for _, tag := range tags {
			if !yield(tag.Key(), tag) {
				return
			}
		}
	})

	// Build the metadata entries, replacing raw EXIF date/time tags with
	// their parsed time.Time equivalent, and dropping the corresponding
	// offset tag once it has been merged in
	entries := make(map[string]gomedia.Metadata, len(tagmap))
	for key, tag := range tagmap {
		entries[key] = tag
	}

	// Replace any tag whose value is a single RATIONAL/SRATIONAL
	// component (e.g. FNumber, ExposureTime, FocalLength, GPSAltitude)
	// with its float64 equivalent
	for key, tag := range tagmap {
		if v, ok := singleRationalToFloat(tag); ok {
			entries[key] = floatMetadata{key: key, value: v, str: tag.Value()}
		}
	}

	for _, dt := range exifDateTimeTags {
		tag, ok := tagmap[dt.dateKey]
		if !ok {
			continue
		}
		var offset string
		if dt.offsetKey != "" {
			if offsetTag, ok := tagmap[dt.offsetKey]; ok {
				offset = offsetTag.Value()
			}
		}
		t, ok := parseExifDateTime(tag.Value(), offset)
		if !ok {
			continue
		}
		entries[dt.dateKey] = timeMetadata{key: dt.dateKey, t: t}
		delete(entries, dt.offsetKey)
	}
	for _, g := range exifGPSTags {
		tag, ok := tagmap[g.dmsKey]
		if !ok {
			continue
		}
		dms, ok := tag.Any().([]libexif.Rational)
		if !ok {
			continue
		}
		deg, ok := dmsToDecimal(dms)
		if !ok {
			continue
		}
		if refTag, ok := tagmap[g.refKey]; ok && strings.EqualFold(refTag.Value(), g.negRef) {
			deg = -deg
		}
		entries[g.dmsKey] = floatMetadata{key: g.dmsKey, value: deg, str: tag.Value()}
		delete(entries, g.refKey)
	}

	// GPSAltitude is negative (below sea level) when GPSAltitudeRef is 1
	if alt, ok := entries["exif:GPSAltitude"].(floatMetadata); ok {
		if refTag, ok := tagmap["exif:GPSAltitudeRef"]; ok {
			if b, ok := refTag.Any().(uint8); ok && b == 1 {
				alt.value = -alt.value
				entries["exif:GPSAltitude"] = alt
			}
			delete(entries, "exif:GPSAltitudeRef")
		}
	}

	return entries
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for jpeg files
	metadata.AddHandler(regexp.MustCompile("^image/jpeg$"), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		// Retrieve the EXIF metadata from the JPEG file
		f, err := exif.Read(r)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		entries := exifTagsToMetadata(f.Tags())
		return metadata.FilterMetadata(entries, filter), nil
	}, "tiff", "exif")
}
