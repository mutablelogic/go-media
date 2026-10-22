package video

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
)

const testDir = "../../etc/test"

// Test_handler_000 checks that video:duration and tag metadata are
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

	meta, err := metadata.GetMetadata(context.Background(), f, contentType)
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

	durVal, ok := got["video:duration"]
	if !ok {
		t.Fatal("expected video:duration in metadata")
	}
	for _, m := range meta {
		if m.Key() != "video:duration" {
			continue
		}
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
	}
	t.Logf("video:duration = %s", durVal)
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
		{"synopsis", "video:synopsis"},
		{"date", "video:year"},
		{"year", "video:year"},
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

// Test_buildVideoEntries_000 checks that "synopsis" becomes dc:description
// when there's no dedicated "description" tag, and isn't also surfaced
// under its own video:synopsis key.
func Test_buildVideoEntries_000(t *testing.T) {
	entries := buildVideoEntries([]gomedia.Metadata{
		meta{key: "synopsis", value: "a synopsis"},
	})
	if _, ok := entries["video:synopsis"]; ok {
		t.Error("did not expect a standalone video:synopsis key")
	}
	if got, ok := entries["dc:description"]; !ok || got.Value() != "a synopsis" {
		t.Errorf("dc:description = %v, want %q", got, "a synopsis")
	}
}

// Test_buildVideoEntries_001 checks that a dedicated "description" tag
// takes precedence over "synopsis", which still doesn't leak under any key.
func Test_buildVideoEntries_001(t *testing.T) {
	entries := buildVideoEntries([]gomedia.Metadata{
		meta{key: "synopsis", value: "a synopsis"},
		meta{key: "description", value: "a description"},
	})
	if got, ok := entries["dc:description"]; !ok || got.Value() != "a description" {
		t.Errorf("dc:description = %v, want %q", got, "a description")
	}
	if _, ok := entries["video:synopsis"]; ok {
		t.Error("did not expect a standalone video:synopsis key")
	}
}

// Test_buildVideoEntries_002 checks that "creation_time" is mirrored to
// dc:date, reformatted as RFC 3339, alongside the original
// video:creation-time key.
func Test_buildVideoEntries_002(t *testing.T) {
	entries := buildVideoEntries([]gomedia.Metadata{
		meta{key: "creation_time", value: "2023-01-15T10:30:00.000000Z"},
	})
	if got, ok := entries["video:creation-time"]; !ok || got.Value() != "2023-01-15T10:30:00.000000Z" {
		t.Errorf("video:creation-time = %v, want the raw tag value", got)
	}
	if got, ok := entries["dc:date"]; !ok || got.Value() != "2023-01-15T10:30:00Z" {
		t.Errorf("dc:date = %v, want %q", got, "2023-01-15T10:30:00Z")
	}
}

// Test_buildVideoEntries_003 checks that an unparseable "creation_time"
// doesn't produce a dc:date entry.
func Test_buildVideoEntries_003(t *testing.T) {
	entries := buildVideoEntries([]gomedia.Metadata{
		meta{key: "creation_time", value: "not-a-timestamp"},
	})
	if _, ok := entries["dc:date"]; ok {
		t.Error("did not expect dc:date for an unparseable creation_time")
	}
}

// Test_buildVideoEntries_quicktime checks that Apple's QuickTime tags are in
// the quicktime namespace, that its creation date, in its own time zone, is
// dc:date rather than creation_time, and that its location is parsed
func Test_buildVideoEntries_quicktime(t *testing.T) {
	entries := buildVideoEntries([]gomedia.Metadata{
		meta{key: "creation_time", value: "2023-03-16T06:23:27.000000Z"},
		meta{key: "com.apple.quicktime.location.accuracy.horizontal", value: "9.424789"},
		meta{key: "com.apple.quicktime.location.ISO6709", value: "+52.4158+013.5896+035.667/"},
		meta{key: "com.apple.quicktime.make", value: "Apple"},
		meta{key: "com.apple.quicktime.model", value: "iPhone 12 mini"},
		meta{key: "com.apple.quicktime.software", value: "16.1.1"},
		meta{key: "com.apple.quicktime.creationdate", value: "2023-03-16T07:23:27+0100"},
	})
	for key, want := range map[string]any{
		"quicktime:make":                         "Apple",
		"quicktime:model":                        "iPhone 12 mini",
		"quicktime:software":                     "16.1.1",
		"quicktime:creationdate":                 "2023-03-16T07:23:27+0100",
		"quicktime:location-iso6709":             "+52.4158+013.5896+035.667/",
		"quicktime:location-accuracy-horizontal": "9.424789",
		"quicktime:latitude":                     52.4158,
		"quicktime:longitude":                    13.5896,
		"quicktime:altitude":                     35.667,
		"dc:date":                                "2023-03-16T07:23:27+01:00",
	} {
		if got, ok := entries[key]; !ok || got.Any() != want {
			t.Errorf("%s = %v, want %v", key, got, want)
		}
	}
	for key := range entries {
		if strings.HasPrefix(key, "video:com-apple") {
			t.Errorf("unexpected %s", key)
		}
	}
}

// Test_buildVideoEntries_language checks that a tag which ffmpeg repeats
// with its language, such as model-nob, is dropped when its value is the
// same, but kept when it isn't
func Test_buildVideoEntries_language(t *testing.T) {
	entries := buildVideoEntries([]gomedia.Metadata{
		meta{key: "model", value: "iPad 2"},
		meta{key: "model-nob", value: "iPad 2"},
		meta{key: "date", value: "2013-03-19T13:46:11+0100"},
		meta{key: "date-nob", value: "2013-03-19T13:46:11+0100"},
		meta{key: "title", value: "Holiday"},
		meta{key: "title-deu", value: "Urlaub"},
	})
	for _, key := range []string{"video:model-nob", "video:date-nob"} {
		if _, ok := entries[key]; ok {
			t.Errorf("did not expect %s", key)
		}
	}
	if got, ok := entries["video:title-deu"]; !ok || got.Value() != "Urlaub" {
		t.Errorf("video:title-deu = %v, want Urlaub", got)
	}
}

// Test_parseISO6709 checks locations in ISO 6709, with and without an
// altitude, and values which aren't locations
func Test_parseISO6709(t *testing.T) {
	if lat, lon, alt, ok := parseISO6709("+51.3218+021.9446+125.949/"); !ok || lat != 51.3218 || lon != 21.9446 || alt == nil || *alt != 125.949 {
		t.Errorf("got %v %v %v %v", lat, lon, alt, ok)
	}
	if lat, lon, alt, ok := parseISO6709("-33.8688+151.2093/"); !ok || lat != -33.8688 || lon != 151.2093 || alt != nil {
		t.Errorf("got %v %v %v %v", lat, lon, alt, ok)
	}
	for _, value := range []string{"", "nowhere", "+95.0+010.0/", "+10.0+190.0/"} {
		if _, _, _, ok := parseISO6709(value); ok {
			t.Errorf("%q parsed", value)
		}
	}
}
