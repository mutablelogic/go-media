package audio

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
)

const testDir = "../../etc/test"

// Test_handler_000 checks that audio:duration is extracted end-to-end via
// GetMetadata for a real audio/* file.
func Test_handler_000(t *testing.T) {
	path := testDir + "/sample.mp3"
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	contentType, _, err := metadata.ContentType(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	meta, err := metadata.GetMetadata(context.Background(), f, contentType)
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, m := range meta {
		if m.Key() != "audio:duration" {
			continue
		}
		found = true
		d, ok := m.Any().(time.Duration)
		if !ok {
			t.Fatalf("Any() = %T, want time.Duration", m.Any())
		}
		if d <= 0 {
			t.Errorf("Duration = %v, want > 0", d)
		}
		sec, err := strconv.ParseFloat(m.Value(), 64)
		if err != nil {
			t.Fatalf("Value() = %q, want numeric seconds: %v", m.Value(), err)
		}
		if sec <= 0 {
			t.Errorf("Value() seconds = %v, want > 0", sec)
		}
		t.Logf("audio:duration = %s (Any()=%v)", m.Value(), d)
	}
	if !found {
		t.Fatal("expected audio:duration in metadata")
	}
}

// Test_sanitizeKey_000 checks that common tag key variants are mapped onto
// their canonical dc:/audio: key.
func Test_sanitizeKey_000(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"title", "dc:title"},
		{"TrackTitle", "dc:title"},
		{"artist", "dc:creator"},
		{"Album_Artist", "dc:creator"},
		{"album", "audio:album"},
		{"AlbumTitle", "audio:album"},
		{"genre", "audio:genre"},
		{"music_genre", "audio:genre"},
		{"year", "audio:year"},
		{"originaldate", "audio:year"},
		{"track", "audio:track"},
		{"iTunes_CDDB_TrackNumber", "audio:track"},
		{"encoder", "audio:encoder"},
	}
	for _, test := range tests {
		if got := sanitizeKey(test.in); got != test.want {
			t.Errorf("sanitizeKey(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}

// Test_sanitizeKey_musicbrainz checks that MusicBrainz's tags, as they're
// spelled in FLAC, MP3 and MP4 files, are mapped onto the musicbrainz
// namespace, named for MusicBrainz's entities
func Test_sanitizeKey_musicbrainz(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"MUSICBRAINZ_ALBUMID", "musicbrainz:release-id"},
		{"MusicBrainz Album Id", "musicbrainz:release-id"},
		{"MUSICBRAINZ_RELEASEGROUPID", "musicbrainz:release-group-id"},
		{"MusicBrainz Release Group Id", "musicbrainz:release-group-id"},
		{"MUSICBRAINZ_TRACKID", "musicbrainz:recording-id"},
		{"MusicBrainz Track Id", "musicbrainz:recording-id"},
		{"MUSICBRAINZ_RELEASETRACKID", "musicbrainz:track-id"},
		{"MusicBrainz Release Track Id", "musicbrainz:track-id"},
		{"MUSICBRAINZ_ARTISTID", "musicbrainz:artist-id"},
		{"MusicBrainz Album Artist Id", "musicbrainz:album-artist-id"},
		{"MUSICBRAINZ_ALBUMARTISTID", "musicbrainz:album-artist-id"},
		{"MusicBrainz Album Type", "musicbrainz:release-type"},
		{"RELEASETYPE", "musicbrainz:release-type"},
		{"MusicBrainz Album Status", "musicbrainz:release-status"},
		{"RELEASESTATUS", "musicbrainz:release-status"},
		{"MusicBrainz Album Release Country", "musicbrainz:release-country"},
		{"RELEASECOUNTRY", "musicbrainz:release-country"},
		{"MusicBrainz Work Id", "musicbrainz:work-id"},
		{"MUSICBRAINZ_NEWTHING", "musicbrainz:newthing"},
		{"musicbrainz", "audio:musicbrainz"},
	}
	for _, test := range tests {
		if got := sanitizeKey(test.in); got != test.want {
			t.Errorf("sanitizeKey(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}

// Test_sanitizeKey_001 checks that noisy/uninteresting tags are dropped.
func Test_sanitizeKey_001(t *testing.T) {
	noise := []string{
		"comment", "iTunNORM", "iTunSMPB", "TAGGING_TIME",
		"major_brand", "minor_version", "compatible_brands",
	}
	for _, key := range noise {
		if got := sanitizeKey(key); got != "" {
			t.Errorf("sanitizeKey(%q) = %q, want \"\" (dropped)", key, got)
		}
	}
}

// Test_handler_playlist checks that a playlist, which has an audio type but
// is text, isn't read by ffmpeg, so has no metadata, rather than an error
func Test_handler_playlist(t *testing.T) {
	for _, contentType := range []string{"audio/x-mpegurl", "audio/mpegurl", "audio/x-scpls"} {
		m3u := strings.NewReader("#EXTM3U\n#EXTINF:123,Artist - Title\nmusic/track01.mp3\n")
		meta, err := metadata.GetMetadata(context.Background(), m3u, contentType, metadata.WithNamespace("audio", "artwork"))
		if err != nil {
			t.Fatalf("%s: %v", contentType, err)
		}
		if len(meta) != 0 {
			t.Fatalf("%s: expected no metadata, got %v", contentType, meta)
		}
	}
}
