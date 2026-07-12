package audio

import (
	"context"
	"os"
	"testing"
	"time"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
)

const testDir = "../../etc/test"

// Test_handler_000 checks that audio:Duration is extracted end-to-end via
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

	meta, err := metadata.GetMetadata(context.Background(), f, contentType, "")
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, m := range meta {
		if m.Key() != "audio:Duration" {
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
		t.Logf("audio:Duration = %s (Any()=%v)", m.Value(), d)
	}
	if !found {
		t.Fatal("expected audio:Duration in metadata")
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
		{"album", "audio:Album"},
		{"AlbumTitle", "audio:Album"},
		{"genre", "audio:Genre"},
		{"music_genre", "audio:Genre"},
		{"year", "audio:Year"},
		{"originaldate", "audio:Year"},
		{"track", "audio:Track"},
		{"iTunes_CDDB_TrackNumber", "audio:Track"},
		{"encoder", "audio:encoder"},
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
