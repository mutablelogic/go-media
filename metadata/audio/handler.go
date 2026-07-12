package audio

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

// meta is a generic gomedia.Metadata for a scalar audio tag value (string,
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
		return v.Truncate(time.Second).String()
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

	// Add metadata handler for audio files
	metadata.AddHandler(regexp.MustCompile(`^audio/.*$`), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		reader, err := ffmpeg.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		entries := make(map[string]gomedia.Metadata)

		// Duration
		entries["audio:Duration"] = meta{key: "audio:Duration", value: reader.Duration()}

		// Tags, normalized and mapped onto dc:/audio: keys where a
		// canonical mapping exists; noisy or uninteresting tags are dropped
		for _, tag := range reader.Metadata() {
			key := sanitizeKey(tag.Key())
			if key == "" {
				continue
			}
			entries[key] = meta{key: key, value: tag.Value()}
		}

		return metadata.FilterMetadata(entries, filter), nil
	}, "dc", "audio")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// sanitizeKey normalizes a raw ffmpeg/format tag key into a
// "namespace:name" metadata key, mapping common variant spellings onto a
// canonical dc:/audio: key, and dropping noisy or uninteresting tags
// (returning "").
func sanitizeKey(key string) string {
	key = strings.ToLower(key)

	// Replace any non-alphanumeric characters with dashes
	key = regexp.MustCompile(`\W+`).ReplaceAllString(key, "-")
	key = strings.ReplaceAll(key, "_", "-")
	key = strings.Trim(key, "-")

	switch key {
	// Noisy or uninteresting tags
	case "eitunnorm", "tagging-time", "accurateripdiscid", "accurateripresult", "comment", "id3v1-comment",
		"id3v2-priv-averagelevel", "id3v2-priv-google-originalclientid", "id3v2-priv-www-amazon-com",
		"itunes-cddb-1", "itunmovi", "itunnorm", "itunsmpb", "gapless-playback", "itunextc",
		"compatible-brands", "itunes-cddb-ids", "account-id", "major-brand", "minor-version":
		return ""

	// Canonical mappings
	case "title", "tracktitle":
		return "dc:title"
	case "artists", "artist", "album-artist":
		return "dc:creator"
	case "album", "albumtitle":
		return "audio:Album"
	case "genre", "music-genre":
		return "audio:Genre"
	case "originalyear", "year", "date", "originaldate", "tdor":
		return "audio:Year"
	case "itunes-cddb-tracknumber", "track", "tracknumber":
		return "audio:Track"
	}

	return "audio:" + key
}
