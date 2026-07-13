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
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
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
	ffmpeg.SetLogging(false, nil)

	// Add metadata handler for video files
	metadata.AddHandler(regexp.MustCompile(`^video/.*$`), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		reader, err := ffmpeg.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		entries := make(map[string]gomedia.Metadata)

		// Duration
		entries["video:Duration"] = meta{key: "video:Duration", value: reader.Duration()}

		// Tags, normalized and mapped onto dc:/video: keys where a
		// canonical mapping exists; noisy or uninteresting tags are dropped
		for _, tag := range reader.Metadata() {
			key := sanitizeKey(tag.Key())
			if key == "" {
				continue
			}
			entries[key] = meta{key: key, value: tag.Value()}
		}

		return metadata.FilterMetadata(entries, filter), nil
	}, "dc", "video")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

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
		return "video:Synopsis"
	case "date", "year":
		return "video:Year"
	}

	return "video:" + key
}
