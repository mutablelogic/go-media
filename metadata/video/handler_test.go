package video

import (
	"context"
	"os"
	"testing"
	"time"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
)

const testDir = "../../etc/test"

// Test_handler_000 checks that video:Duration and tag metadata are
// extracted end-to-end via GetMetadata for a real video/* file.
func Test_handler_000(t *testing.T) {
	path := testDir + "/sample.mp4"
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

	got := make(map[string]string, len(meta))
	for _, m := range meta {
		got[m.Key()] = m.Value()
	}

	if got["dc:title"] != "Sample From Big Buck Bunny" {
		t.Errorf("dc:title = %q, want %q", got["dc:title"], "Sample From Big Buck Bunny")
	}

	durVal, ok := got["video:Duration"]
	if !ok {
		t.Fatal("expected video:Duration in metadata")
	}
	for _, m := range meta {
		if m.Key() != "video:Duration" {
			continue
		}
		d, ok := m.Any().(time.Duration)
		if !ok {
			t.Fatalf("Any() = %T, want time.Duration", m.Any())
		}
		if d <= 0 {
			t.Errorf("Duration = %v, want > 0", d)
		}
	}
	t.Logf("video:Duration = %s", durVal)
}

// Test_sanitizeKey_000 checks that common tag key variants are mapped onto
// their canonical dc:/video: key.
func Test_sanitizeKey_000(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"title", "dc:title"},
		{"director", "dc:creator"},
		{"description", "dc:description"},
		{"synopsis", "video:Synopsis"},
		{"date", "video:Year"},
		{"year", "video:Year"},
		{"encoder", "video:encoder"},
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
		"comment", "major_brand", "minor_version", "compatible_brands",
		"iTunes_CDDB_1", "iTunMOVI", "gapless_playback", "iTunEXTC",
	}
	for _, key := range noise {
		if got := sanitizeKey(key); got != "" {
			t.Errorf("sanitizeKey(%q) = %q, want \"\" (dropped)", key, got)
		}
	}
}
