package video

import (
	"context"
	"fmt"
	"image"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// meta is a generic gomedia.Metadata for a scalar video tag value (string,
// time.Duration, or float64).
type meta struct {
	key   string
	value any
}

func (m meta) Key() string        { return m.key }
func (m meta) Bytes() []byte      { return nil }
func (m meta) Image() image.Image { return nil }
func (m meta) Any() any           { return m.value }

func (m meta) Value() string {
	switch v := m.value.(type) {
	case string:
		return v
	case time.Duration:
		return strconv.FormatFloat(v.Seconds(), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Suppress ffmpeg's own logging
	ff.AVUtil_log_set_level(ff.AV_LOG_ERROR)

	// Add metadata handler for video files
	metadata.AddHandler(regexp.MustCompile(`^video/.*$`), "ffmpeg", func(_ context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		// A playlist, which is text, isn't read by ffmpeg
		if o.IsText() {
			return nil, nil
		}

		rd, err := reader.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer rd.Close()

		entries := buildVideoEntries(rd.Metadata())
		entries["video:duration"] = meta{key: "video:duration", value: rd.Duration()}

		return metadata.FilterMetadata(entries, o), nil
	}, "dc", "video", "quicktime")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// buildVideoEntries normalizes a container's raw tags and maps them onto
// dc:/video: keys where a canonical mapping exists (see sanitizeKey),
// dropping noisy or uninteresting tags. It's kept separate from the
// ffmpeg-backed handler so the merge logic below - which needs to see
// every tag before deciding some of the output - can be tested without a
// real media file:
//
//   - "synopsis" only becomes dc:description when there's no dedicated
//     "description" tag to prefer instead; it's never surfaced as its own
//     video:synopsis key.
//   - "creation-time" is additionally mirrored to dc:date, reformatted as
//     RFC 3339, whenever it parses as a timestamp, unless there's a
//     QuickTime creation date, which is when the video was recorded, in its
//     own time zone, whereas creation-time is in UTC, and can be when the
//     file was written.
//   - A QuickTime location, in ISO 6709, is additionally parsed into
//     quicktime:latitude, quicktime:longitude and quicktime:altitude.
func buildVideoEntries(tags []gomedia.Metadata) map[string]gomedia.Metadata {
	entries := make(map[string]gomedia.Metadata)

	// The values of the tags, by their names, so a tag repeated with its
	// language, such as model-eng for model, which has the same value, is
	// dropped
	values := make(map[string]string, len(tags))
	for _, tag := range tags {
		values[strings.ToLower(tag.Key())] = tag.Value()
	}

	var synopsis, creationTime, creationDate, location string
	for _, tag := range tags {
		key := sanitizeKey(tag.Key())
		if key == "" {
			continue
		} else if match := languageSuffix.FindStringSubmatch(strings.ToLower(tag.Key())); match != nil {
			if value, exists := values[match[1]]; exists && value == tag.Value() {
				continue
			}
		}
		switch key {
		case "video:synopsis":
			synopsis = tag.Value()
			continue
		case "video:creation-time":
			creationTime = tag.Value()
		case "quicktime:creationdate":
			creationDate = tag.Value()
		case "quicktime:location-iso6709":
			location = tag.Value()
		}
		entries[key] = meta{key: key, value: tag.Value()}
	}

	if _, ok := entries["dc:description"]; !ok && synopsis != "" {
		entries["dc:description"] = meta{key: "dc:description", value: synopsis}
	}
	if t, err := time.Parse(quicktimeDate, creationDate); err == nil {
		entries["dc:date"] = meta{key: "dc:date", value: t.Format(time.RFC3339)}
	} else if t, err := time.Parse(time.RFC3339Nano, creationTime); err == nil {
		entries["dc:date"] = meta{key: "dc:date", value: t.Format(time.RFC3339)}
	}
	if latitude, longitude, altitude, ok := parseISO6709(location); ok {
		entries["quicktime:latitude"] = meta{key: "quicktime:latitude", value: latitude}
		entries["quicktime:longitude"] = meta{key: "quicktime:longitude", value: longitude}
		if altitude != nil {
			entries["quicktime:altitude"] = meta{key: "quicktime:altitude", value: *altitude}
		}
	}

	return entries
}

// languageSuffix matches the name of a tag with a language, such as
// model-eng, which ffmpeg adds to a tag with a language
var languageSuffix = regexp.MustCompile(`^(.+)-[a-z]{3}$`)

// quicktimeDate is the layout of a QuickTime creation date, such as
// 2023-03-16T07:23:27+0100
const quicktimeDate = "2006-01-02T15:04:05-0700"

// iso6709 matches a location in ISO 6709, in decimal degrees, such as
// +52.4158+013.5896+035.667/, whose altitude, in metres, is optional
var iso6709 = regexp.MustCompile(`^([+-]\d{1,2}(?:\.\d+)?)([+-]\d{1,3}(?:\.\d+)?)([+-]\d+(?:\.\d+)?)?(?:CRS[^/]*)?/?$`)

// parseISO6709 returns the latitude, longitude and altitude, which is nil
// when it isn't known, of a location in ISO 6709, in decimal degrees, or
// false when it isn't one
func parseISO6709(value string) (float64, float64, *float64, bool) {
	match := iso6709.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return 0, 0, nil, false
	}
	latitude, err := strconv.ParseFloat(match[1], 64)
	if err != nil || latitude < -90 || latitude > 90 {
		return 0, 0, nil, false
	}
	longitude, err := strconv.ParseFloat(match[2], 64)
	if err != nil || longitude < -180 || longitude > 180 {
		return 0, 0, nil, false
	}
	if match[3] == "" {
		return latitude, longitude, nil, true
	}
	altitude, err := strconv.ParseFloat(match[3], 64)
	if err != nil {
		return latitude, longitude, nil, true
	}
	return latitude, longitude, &altitude, true
}

// sanitizeKey normalizes a raw ffmpeg/format tag key into a
// "namespace:name" metadata key, mapping common variant spellings onto a
// canonical dc:/video: key, and dropping noisy or uninteresting tags
// (returning "").
func sanitizeKey(key string) string {
	key = strings.ToLower(key)

	// Replace any non-alphanumeric characters with dashes
	key = regexp.MustCompile(`\W+`).ReplaceAllString(key, "-")
	key = strings.ReplaceAll(key, "_", "-")
	key = strings.Trim(key, "-")

	// Apple's QuickTime tags, such as com.apple.quicktime.model, are in the
	// quicktime namespace
	if name, found := strings.CutPrefix(key, "com-apple-quicktime-"); found && name != "" {
		return "quicktime:" + name
	}

	switch key {
	// Noisy or uninteresting tags
	case "compatible-brands", "major-brand", "minor-version", "comment",
		"itunes-cddb-1", "itunmovi", "gapless-playback", "itunextc":
		return ""

	// Canonical mappings
	case "title":
		return "dc:title"
	case "director":
		return "dc:creator"
	case "description":
		return "dc:description"
	case "synopsis":
		return "video:synopsis"
	case "date", "year":
		return "video:year"
	}

	return "video:" + key
}
