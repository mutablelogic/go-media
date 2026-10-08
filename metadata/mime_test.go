package metadata_test

import (
	"bytes"
	"mime"
	"os"
	"path/filepath"
	"testing"

	// Packages
	. "github.com/mutablelogic/go-media/metadata"
)

const (
	TEST_DIR = "../etc/test"
)

type namedReader struct {
	*bytes.Reader
	name string
}

func (r namedReader) Name() string {
	return r.name
}

func Test_mime_000(t *testing.T) {
	entries, err := os.ReadDir(TEST_DIR)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			f, err := os.Open(filepath.Join(TEST_DIR, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			contentType, meta, err := ContentType(f)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s: %s %v", entry.Name(), contentType, meta)
		})
	}
}

func Test_mime_001_m4a_override(t *testing.T) {
	// MP4 family signature that usually sniffs as video/mp4.
	data := []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}
	r := namedReader{Reader: bytes.NewReader(data), name: "audio.m4a"}

	contentType, _, err := ContentType(r)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "audio/mp4" {
		t.Fatalf("expected audio/mp4, got %q", contentType)
	}
}

func Test_mime_002_short_reader_uses_read_length(t *testing.T) {
	r := namedReader{Reader: bytes.NewReader([]byte("hello")), name: "sample.txt"}

	contentType, _, err := ContentType(r)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "text/plain" {
		t.Fatalf("expected text/plain, got %q", contentType)
	}
}

func Test_mime_003_extensionByType(t *testing.T) {
	if got := ExtensionByType("image/jpeg"); got != ".jpg" {
		t.Fatalf("ExtensionByType(image/jpeg) = %q, want %q", got, ".jpg")
	}
	if got := ExtensionByType("application/x-not-a-real-type"); got != "" {
		t.Fatalf("ExtensionByType(unknown) = %q, want empty", got)
	}
}

// unreadableReader fails the test if Read is ever called on it.
type unreadableReader struct {
	t    *testing.T
	name string
}

func (r unreadableReader) Name() string { return r.name }

func (r unreadableReader) Read([]byte) (int, error) {
	r.t.Fatal("Read was called; curated extensions should return without reading the stream")
	return 0, nil
}

func Test_mime_004_curated_extension_skips_read(t *testing.T) {
	r := unreadableReader{t: t, name: "audio.m4a"}

	contentType, _, err := ContentType(r)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "audio/mp4" {
		t.Fatalf("expected audio/mp4, got %q", contentType)
	}
}

func Test_mime_005_jpg_extension_is_not_authoritative(t *testing.T) {
	// PNG signature in a file misleadingly named .jpg - sniffing must win,
	// proving .jpg (Preferred for ExtensionByType, but not Authoritative)
	// doesn't skip reading the stream the way .m4a deliberately does.
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	r := namedReader{Reader: bytes.NewReader(data), name: "mislabeled.jpg"}

	contentType, _, err := ContentType(r)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "image/png" {
		t.Fatalf("expected image/png (from sniffing, not the .jpg extension), got %q", contentType)
	}
}

func Test_mime_006_text_extensions(t *testing.T) {
	// Text sniffs as text/plain, so the extension says what kind of text it
	// is, keeping the charset from sniffing
	for name, want := range map[string]string{
		"main.go":       "text/x-go",
		"notes.nfo":     "text/x-nfo",
		"movie.srt":     "application/x-subrip",
		"config.yaml":   "application/yaml",
		"app.ts":        "text/typescript",
		"plain.txt":     "text/plain",
		"unknown.xyz12": "text/plain",
		// Playlists have audio types, which would be read by the audio
		// handlers, so their text content stays plain text
		"playlist.m3u": "text/plain",
		"playlist.pls": "text/plain",
	} {
		r := namedReader{Reader: bytes.NewReader([]byte("package main\n\nfunc main() {}\n")), name: name}
		contentType, params, err := ContentType(r)
		if err != nil {
			t.Fatal(name, err)
		}
		if contentType != want {
			t.Errorf("%s: expected %q, got %q", name, want, contentType)
		}
		if params["charset"] != "utf-8" {
			t.Errorf("%s: expected charset utf-8, got %v", name, params)
		}
	}
}

func Test_mime_007_ts_transport_stream(t *testing.T) {
	// A .ts file which isn't text is an MPEG transport stream, whose
	// 188-byte packets start with 0x47
	data := make([]byte, 376)
	data[0], data[188] = 0x47, 0x47
	for name, r := range map[string]namedReader{
		"recording.ts": {Reader: bytes.NewReader(data), name: "recording.ts"},
		"empty.ts":     {Reader: bytes.NewReader(nil), name: "empty.ts"},
	} {
		contentType, _, err := ContentType(r)
		if err != nil {
			t.Fatal(name, err)
		}
		if contentType != "video/mp2t" {
			t.Errorf("%s: expected video/mp2t, got %q", name, contentType)
		}
	}
}

func Test_mime_008_containers_are_authoritative(t *testing.T) {
	// Containers which sniff as another type are known by their extension,
	// without reading the stream
	for name, want := range map[string]string{
		"movie.mkv":  "video/x-matroska",
		"MOVIE.MKV":  "video/x-matroska",
		"audio.mka":  "audio/x-matroska",
		"clip.mov":   "video/quicktime",
		"clip.m4v":   "video/x-m4v",
		"track.opus": "audio/opus",
		"book.epub":  "application/epub+zip",
	} {
		contentType, _, err := ContentType(unreadableReader{t: t, name: name})
		if err != nil {
			t.Fatal(name, err)
		}
		if contentType != want {
			t.Errorf("%s: expected %q, got %q", name, want, contentType)
		}
	}
}

func Test_mime_009_registered(t *testing.T) {
	// The types are registered with the mime package, on every platform
	for ext, want := range map[string]string{
		".mkv":  "video/x-matroska",
		".nfo":  "text/x-nfo; charset=utf-8",
		".ts":   "video/mp2t",
		".go":   "text/x-go; charset=utf-8",
		".epub": "application/epub+zip",
	} {
		if got := mime.TypeByExtension(ext); got != want {
			t.Errorf("TypeByExtension(%s) = %q, want %q", ext, got, want)
		}
	}
	for contentType, want := range map[string]string{
		"video/x-matroska": ".mkv",
		"video/mp2t":       ".ts",
		"audio/ogg":        ".ogg",
	} {
		if got := ExtensionByType(contentType); got != want {
			t.Errorf("ExtensionByType(%s) = %q, want %q", contentType, got, want)
		}
	}
}
