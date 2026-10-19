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

// The extensions which are refined are registered by the caller, rather
// than being known to the package
func init() {
	for ext, contentType := range map[string]string{
		".md":       "text/markdown",
		".markdown": "text/markdown",
		".docx":     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx":     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx":     "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".odt":      "application/vnd.oasis.opendocument.text",
		".epub":     "application/epub+zip",
		".gdoc":     "application/vnd.google-apps.document",
	} {
		if err := mime.AddExtensionType(ext, contentType); err != nil {
			panic(err)
		}
	}
}

func Test_mime_006_text_refined_by_extension(t *testing.T) {
	// Markdown sniffs as text/plain, and its extension refines it, keeping
	// the sniffed charset
	for _, name := range []string{"README.md", "notes.MD", "doc.markdown"} {
		r := namedReader{Reader: bytes.NewReader([]byte("# Title\n\nSome text\n")), name: name}
		contentType, params, err := ContentType(r)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != "text/markdown" {
			t.Fatalf("%s: expected text/markdown, got %q", name, contentType)
		}
		if params["charset"] != "utf-8" {
			t.Fatalf("%s: expected charset utf-8, got %v", name, params)
		}
	}
}

func Test_mime_007_text_extension_is_not_authoritative(t *testing.T) {
	// Binary content named .md is detected from its bytes
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	r := namedReader{Reader: bytes.NewReader(data), name: "image.md"}
	contentType, _, err := ContentType(r)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "image/png" {
		t.Fatalf("expected image/png, got %q", contentType)
	}

	// Plain text without a more specific extension is text/plain
	r = namedReader{Reader: bytes.NewReader([]byte("hello")), name: "notes.txt"}
	if contentType, _, err := ContentType(r); err != nil {
		t.Fatal(err)
	} else if contentType != "text/plain" {
		t.Fatalf("expected text/plain, got %q", contentType)
	}
}

func Test_mime_008_zip_refined_by_extension(t *testing.T) {
	// A document which is a zip archive sniffs as application/zip, and its
	// extension refines it
	zip := []byte("PK\x03\x04\x14\x00\x00\x00\x08\x00")
	for name, want := range map[string]string{
		"report.docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"sheet.XLSX":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"deck.pptx":   "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"letter.odt":  "application/vnd.oasis.opendocument.text",
		"book.epub":   "application/epub+zip",
		"archive.zip": "application/zip",
		"image.jpg":   "application/zip",
	} {
		r := namedReader{Reader: bytes.NewReader(zip), name: name}
		contentType, _, err := ContentType(r)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != want {
			t.Fatalf("%s: expected %q, got %q", name, want, contentType)
		}
	}
}

func Test_mime_009_text_refined_by_application_type(t *testing.T) {
	// Plain text is refined by an application type which is text, such as a
	// registered type, but not by a binary type
	for name, want := range map[string]string{
		"Bader Pass.gdoc": "application/vnd.google-apps.document",
		"data.json":       "application/json",
		"image.png":       "text/plain",
		"report.docx":     "text/plain",
		"notes.txt":       "text/plain",
	} {
		r := namedReader{Reader: bytes.NewReader([]byte(`{"doc_id":"abc"}`)), name: name}
		contentType, params, err := ContentType(r)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != want {
			t.Fatalf("%s: expected %q, got %q", name, want, contentType)
		}
		if params["charset"] != "utf-8" {
			t.Fatalf("%s: expected charset utf-8, got %v", name, params)
		}
	}
}

func Test_mime_010_svg_refined_by_extension(t *testing.T) {
	// An SVG image sniffs as XML with a declaration, as HTML with a comment,
	// and as plain text otherwise, and its extension refines it
	for _, data := range []string{
		`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<svg xmlns="http://www.w3.org/2000/svg"/>`,
		`<?xml version="1.0"?>` + "\n" + `<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">` + "\n" + `<svg xmlns="http://www.w3.org/2000/svg"/>`,
		`<!-- Generator: Adobe Illustrator -->` + "\n" + `<svg xmlns="http://www.w3.org/2000/svg"/>`,
		`<svg xmlns="http://www.w3.org/2000/svg"/>`,
	} {
		r := namedReader{Reader: bytes.NewReader([]byte(data)), name: "logo.svg"}
		contentType, _, err := ContentType(r)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != "image/svg+xml" {
			t.Fatalf("%q: expected image/svg+xml, got %q", data, contentType)
		}
	}

	// A PNG image named .svg is detected from its bytes
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	r := namedReader{Reader: bytes.NewReader(png), name: "logo.svg"}
	if contentType, _, err := ContentType(r); err != nil {
		t.Fatal(err)
	} else if contentType != "image/png" {
		t.Fatalf("expected image/png, got %q", contentType)
	}
}
