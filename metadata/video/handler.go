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
	metadata.AddHandler(regexp.MustCompile(`^video/.*$`), func(_ context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		rd, err := reader.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer rd.Close()

		entries := buildVideoEntries(rd.Metadata())
		entries["video:duration"] = meta{key: "video:duration", value: rd.Duration()}

		return metadata.FilterMetadata(entries, o), nil
	}, "dc", "video")
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
//     RFC 3339, whenever it parses as a timestamp.
func buildVideoEntries(tags []gomedia.Metadata) map[string]gomedia.Metadata {
	entries := make(map[string]gomedia.Metadata)

	var synopsis, creationTime string
	for _, tag := range tags {
		key := sanitizeKey(tag.Key())
		if key == "" {
			continue
		}
		switch key {
		case "video:synopsis":
			synopsis = tag.Value()
			continue
		case "video:creation-time":
			creationTime = tag.Value()
		}
		entries[key] = meta{key: key, value: tag.Value()}
	}

	if _, ok := entries["dc:description"]; !ok && synopsis != "" {
		entries["dc:description"] = meta{key: "dc:description", value: synopsis}
	}
	if t, err := time.Parse(time.RFC3339Nano, creationTime); err == nil {
		entries["dc:date"] = meta{key: "dc:date", value: t.Format(time.RFC3339)}
	}

	return entries
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
