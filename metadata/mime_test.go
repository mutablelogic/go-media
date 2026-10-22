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

		// As on Linux, where an icon's extension is the other type of icon
		".ico": "image/vnd.microsoft.icon",
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

func Test_mime_011_tga_refined_by_extension(t *testing.T) {
	// An uncompressed TGA image starts with the same four bytes as an icon,
	// so its extension refines it, but an icon's doesn't
	tga := []byte{0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x20, 0x08}
	for name, want := range map[string]string{
		"steam_quad.tga": "image/x-tga",
		"STEAM_QUAD.TGA": "image/x-tga",
		"favicon.ico":    "image/x-icon",
		"notes.txt":      "image/x-icon",
	} {
		r := namedReader{Reader: bytes.NewReader(tga), name: name}
		contentType, _, err := ContentType(r)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != want {
			t.Fatalf("%s: expected %q, got %q", name, want, contentType)
		}
	}
}

func Test_mime_012_playlist_refined_by_extension(t *testing.T) {
	// A playlist sniffs as plain text, and its extension refines it, with
	// the same type on every platform
	m3u := []byte("#EXTM3U\n#EXTINF:123,Artist - Title\nmusic/track01.mp3\n")
	for name, want := range map[string]string{
		"mix.m3u":   "audio/x-mpegurl",
		"MIX.M3U":   "audio/x-mpegurl",
		"live.m3u8": "application/vnd.apple.mpegurl",
		"radio.pls": "audio/x-scpls",
		"notes.txt": "text/plain",
	} {
		r := namedReader{Reader: bytes.NewReader(m3u), name: name}
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

	// A binary file named .m3u is detected from its bytes
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	r := namedReader{Reader: bytes.NewReader(png), name: "mix.m3u"}
	if contentType, _, err := ContentType(r); err != nil {
		t.Fatal(err)
	} else if contentType != "image/png" {
		t.Fatalf("expected image/png, got %q", contentType)
	}
}

func Test_mime_013_tga_extension_by_type(t *testing.T) {
	// The extension of a TGA image is known on every platform
	if ext := ExtensionByType("image/x-tga"); ext != ".tga" {
		t.Fatalf("expected .tga, got %q", ext)
	}
}

func Test_mime_014_m4p_override(t *testing.T) {
	// Protected audio from iTunes or Apple Music is an MP4 file, which sniffs
	// as video/mp4, so its extension decides, as with .m4a
	data := []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p', 'M', '4', 'A', ' ', 0x00, 0x00, 0x00, 0x00, 'M', '4', 'A', ' ', 'm', 'p', '4', '2', 'i', 's', 'o', 'm'}
	for _, name := range []string{"04 Jungle Love.m4p", "JUMP.M4P"} {
		r := namedReader{Reader: bytes.NewReader(data), name: name}
		contentType, _, err := ContentType(r)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != "audio/mp4" {
			t.Fatalf("%s: expected audio/mp4, got %q", name, contentType)
		}
	}

	// audio/mp4's preferred extension is still .m4a
	if ext := ExtensionByType("audio/mp4"); ext != ".m4a" {
		t.Fatalf("expected .m4a, got %q", ext)
	}
}

func Test_mime_015_text_extension_type(t *testing.T) {
	// The extension is an MPEG transport stream's, whatever the host's MIME
	// types are, and TypeScript's when it's text
	if err := mime.AddExtensionType(".ts", "video/mp2t"); err != nil {
		t.Fatal(err)
	}
	if err := AddTextExtensionType(".ts", "text/x-typescript; charset=utf-8"); err != nil {
		t.Fatal(err)
	}

	// TypeScript sniffs as plain text, and is refined by the extension's
	// type of text, keeping its charset, whatever the case of its extension
	for _, name := range []string{"index.ts", "INDEX.TS"} {
		r := namedReader{Reader: bytes.NewReader([]byte("export const answer: number = 42;\n")), name: name}
		contentType, params, err := ContentType(r)
		if err != nil {
			t.Fatal(err)
		}
		if contentType != "text/x-typescript" {
			t.Fatalf("%s: expected text/x-typescript, got %q", name, contentType)
		}
		if params["charset"] != "utf-8" {
			t.Fatalf("%s: expected charset utf-8, got %v", name, params)
		}
	}

	// An MPEG transport stream, which isn't text, has the type it otherwise
	// has, which is its extension's, rather than the type of text
	packet := make([]byte, 188*3)
	for i := 0; i < len(packet); i += 188 {
		packet[i] = 0x47
		packet[i+1] = 0x40
	}
	r := namedReader{Reader: bytes.NewReader(packet), name: "recording.ts"}
	contentType, _, err := ContentType(r)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "video/mp2t" {
		t.Fatalf("expected video/mp2t, got %q", contentType)
	}

	// The extension of the type of text is its registered extension, without
	// the type's parameters or case
	for _, contentType := range []string{"text/x-typescript", "text/x-typescript; charset=utf-8", "Text/X-TypeScript"} {
		if ext := ExtensionByType(contentType); ext != ".ts" {
			t.Fatalf("%q: expected .ts, got %q", contentType, ext)
		}
	}
}

func Test_mime_015_text_extension_type_html(t *testing.T) {
	// A Vue or Svelte component, which starts with a <script> tag or a
	// comment, sniffs as HTML, and is refined by its extension's type of
	// text, while HTML is HTML
	if err := AddTextExtensionType(".vue", "text/x-vue"); err != nil {
		t.Fatal(err)
	}
	if err := AddTextExtensionType(".svelte", "text/x-svelte"); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"App.vue":       "text/x-vue",
		"Button.svelte": "text/x-svelte",
		"index.html":    "text/html",
	} {
		for _, content := range []string{"<script setup>\nconst a = 1\n</script>\n", "<!-- A component -->\n<template><p>Hi</p></template>\n"} {
			r := namedReader{Reader: bytes.NewReader([]byte(content)), name: name}
			contentType, params, err := ContentType(r)
			if err != nil {
				t.Fatal(err)
			}
			if contentType != want || params["charset"] != "utf-8" {
				t.Errorf("%s %q: got %q %v, want %q", name, content, contentType, params, want)
			}
		}
	}
}

func Test_mime_016_text_extension_type_invalid(t *testing.T) {
	for _, tt := range []struct{ ext, contentType string }{
		{"ts", "text/x-typescript"},
		{".", "text/x-typescript"},
		{".ts", "not a type"},
	} {
		if err := AddTextExtensionType(tt.ext, tt.contentType); err == nil {
			t.Fatalf("%q, %q: expected an error", tt.ext, tt.contentType)
		}
	}
}
