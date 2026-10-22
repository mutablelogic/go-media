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
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
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
		return strconv.FormatFloat(v.Seconds(), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case []string:
		return strings.Join(v, ", ")
	default:
		return fmt.Sprint(v)
	}
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Suppress ffmpeg's own logging
	ff.AVUtil_log_set_level(ff.AV_LOG_ERROR)

	// Add metadata handler for audio files
	metadata.AddHandler(regexp.MustCompile(`^audio/.*$`), "ffmpeg", func(_ context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		// A playlist, which is text, isn't read by ffmpeg
		if o.IsText() {
			return nil, nil
		}

		rd, err := reader.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer rd.Close()

		entries := make(map[string]gomedia.Metadata)

		// Duration
		entries["audio:duration"] = meta{key: "audio:duration", value: rd.Duration()}

		// Tags, normalized and mapped onto dc:/audio: keys where a
		// canonical mapping exists; noisy or uninteresting tags are dropped
		for _, tag := range rd.Metadata() {
			key := sanitizeKey(tag.Key())
			if key == "" {
				continue
			}
			entries[key] = meta{key: key, value: tag.Value()}
		}
		fillFromSortTags(entries)

		return metadata.FilterMetadata(entries, o), nil
	}, "dc", "audio", "musicbrainz")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// sortTags are the keys which a sort tag fills in, when they're missing
var sortTags = []struct{ key, sort string }{
	{"dc:title", "audio:sort-name"},
	{"dc:creator", "audio:sort-artist"},
	{"dc:creator", "audio:sort-album-artist"},
	{"audio:album", "audio:sort-album"},
}

// fillFromSortTags fills in the title, artist and album from the tags which
// are for sorting, such as sort_name, when the file doesn't have them. Music
// downloaded from Apple Music only has the sort tags.
func fillFromSortTags(entries map[string]gomedia.Metadata) {
	for _, tag := range sortTags {
		if _, exists := entries[tag.key]; exists {
			continue
		}
		if sorted, exists := entries[tag.sort]; exists && strings.TrimSpace(sorted.Value()) != "" {
			entries[tag.key] = meta{key: tag.key, value: sorted.Value()}
		}
	}
}

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
		return "audio:album"
	case "genre", "music-genre":
		return "audio:genre"
	case "originalyear", "year", "date", "originaldate", "tdor":
		return "audio:year"
	case "itunes-cddb-tracknumber", "track", "tracknumber":
		return "audio:track"
	}

	// MusicBrainz's identifiers, and the details of the release
	if key, found := musicbrainzKey(key); found {
		return key
	}
	return "audio:" + key
}

// musicbrainzTags are the keys of MusicBrainz's tags, by their names without
// separators, which are named for MusicBrainz's entities: an album is a
// release, and a track is a recording, and a release's track is a track.
// Tags are spelled differently in each container, such as
// MUSICBRAINZ_ALBUMID in a FLAC file, and MusicBrainz Album Id in an MP3 or
// MP4 file, and the details of a release don't have the prefix in a FLAC
// file, such as RELEASETYPE for MusicBrainz Album Type.
var musicbrainzTags = map[string]string{
	"albumid":             "musicbrainz:release-id",
	"releasegroupid":      "musicbrainz:release-group-id",
	"trackid":             "musicbrainz:recording-id",
	"releasetrackid":      "musicbrainz:track-id",
	"artistid":            "musicbrainz:artist-id",
	"albumartistid":       "musicbrainz:album-artist-id",
	"workid":              "musicbrainz:work-id",
	"discid":              "musicbrainz:disc-id",
	"originalalbumid":     "musicbrainz:original-release-id",
	"originalartistid":    "musicbrainz:original-artist-id",
	"albumtype":           "musicbrainz:release-type",
	"albumstatus":         "musicbrainz:release-status",
	"albumreleasecountry": "musicbrainz:release-country",
}

// musicbrainzDetails are the details of a release in a FLAC file, which
// don't have MusicBrainz's prefix, by their names
var musicbrainzDetails = map[string]string{
	"releasetype":    "musicbrainz:release-type",
	"releasestatus":  "musicbrainz:release-status",
	"releasecountry": "musicbrainz:release-country",
}

// musicbrainzKey returns the key of a MusicBrainz tag, whose key is
// normalized, or false when it isn't one. A tag with MusicBrainz's prefix
// which isn't known keeps its name.
func musicbrainzKey(key string) (string, bool) {
	if key, found := musicbrainzDetails[key]; found {
		return key, true
	}
	name, found := strings.CutPrefix(key, "musicbrainz-")
	if !found || name == "" {
		return "", false
	} else if key, found := musicbrainzTags[strings.ReplaceAll(name, "-", "")]; found {
		return key, true
	}
	return "musicbrainz:" + name, true
}
