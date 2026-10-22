package subtitle_test

import (
	"context"
	"strings"
	"testing"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	_ "github.com/mutablelogic/go-media/metadata/subtitle"
)

// namedReader is a reader with a file name
type namedReader struct {
	*strings.Reader
	name string
}

func (r namedReader) Name() string { return r.name }

// srt has a BOM, CRLF line endings, styling, a repeated cue, a cue without a
// blank line after it, Markdown's characters, and a line in Windows-1252
const srt = "\uFEFF1\r\n00:00:01,000 --> 00:00:03,500\r\n<i>Hello</i> there.\r\n{\\an8}General Kenobi!\r\n\r\n" +
	"2\r\n00:00:04,000 --> 00:00:05,000\r\nHello there.\r\nGeneral Kenobi!\r\n\r\n" +
	"3\r\n00:01:02.5 --> 00:01:04,000\r\n*Not* bold, [not] a link\r\n" +
	"4\r\n01:02:03,000 --> 01:02:05,000\r\nCaf\xe9 au lait\r\n\r\n" +
	"5\r\n01:02:06,000 --> 01:02:07,000\r\n<font color=\"#fff\"></font>\r\n"

const srtMarkdown = "**1 · 00:00:01** Hello there.\nGeneral Kenobi!\n\n" +
	"**3 · 00:01:02** \\*Not\\* bold, \\[not\\] a link\n\n" +
	"**4 · 01:02:03** Café au lait\n"

// markdown returns the body and next of the subtitles' Markdown
func markdown(t *testing.T, opts ...metadata.Option) (string, string) {
	t.Helper()
	r := namedReader{strings.NewReader(srt), "film.srt"}
	contentType, _, err := metadata.ContentType(r)
	if err != nil {
		t.Fatal(err)
	} else if contentType != "application/x-subrip" {
		t.Fatalf("content type is %q", contentType)
	}
	meta, err := metadata.GetMetadata(context.Background(), namedReader{strings.NewReader(srt), "film.srt"}, contentType, append(opts, metadata.WithNamespace(gomedia.MetaMarkdown))...)
	if err != nil {
		t.Fatal(err)
	}
	var body, next string
	for _, entry := range meta {
		switch entry.Key() {
		case "markdown:body":
			body = entry.Value()
		case "markdown:next":
			next = entry.Value()
		}
	}
	return body, next
}

func Test_srt_000(t *testing.T) {
	// The cues are paragraphs, which start with their times, without their
	// numbers, styling, or the repeated cue, and with Markdown's characters
	// escaped
	if body, next := markdown(t); body != srtMarkdown || next != "" {
		t.Errorf("got %q, next %q", body, next)
	}
}

func Test_srt_001(t *testing.T) {
	// A part of the Markdown, and the next
	body, next := markdown(t, metadata.WithRange(0, 12))
	if body != srtMarkdown[:12] || next != "12" {
		t.Errorf("got %q, next %q", body, next)
	}
}

func Test_srt_003(t *testing.T) {
	// UTF-16 subtitles, with a byte order mark, in either byte order
	for _, bigEndian := range []bool{false, true} {
		data := []byte{0xFF, 0xFE}
		if bigEndian {
			data = []byte{0xFE, 0xFF}
		}
		for _, r := range "1\r\n00:00:01,000 --> 00:00:02,000\r\nCafé\r\n" {
			if bigEndian {
				data = append(data, byte(r>>8), byte(r))
			} else {
				data = append(data, byte(r), byte(r>>8))
			}
		}
		meta, err := metadata.GetMetadata(context.Background(), strings.NewReader(string(data)), "application/x-subrip", metadata.WithNamespace(gomedia.MetaMarkdown))
		if err != nil || len(meta) != 1 || meta[0].Value() != "**1 · 00:00:01** Café\n" {
			t.Errorf("big endian %v: got %v, %v", bigEndian, meta, err)
		}
	}
}

func Test_srt_004(t *testing.T) {
	// A cue without a number is the previous cue's and one, and the first
	// is one
	const unnumbered = "00:00:01,000 --> 00:00:02,000\nOne\n\n00:00:03,000 --> 00:00:04,000\nTwo\n\n7\n00:00:05,000 --> 00:00:06,000\nSeven\n\n00:00:07,000 --> 00:00:08,000\nEight\n"
	meta, err := metadata.GetMetadata(context.Background(), strings.NewReader(unnumbered), "application/x-subrip", metadata.WithNamespace(gomedia.MetaMarkdown))
	want := "**1 · 00:00:01** One\n\n**2 · 00:00:03** Two\n\n**7 · 00:00:05** Seven\n\n**8 · 00:00:07** Eight\n"
	if err != nil || len(meta) != 1 || meta[0].Value() != want {
		t.Errorf("got %v, %v", meta, err)
	}
}

func Test_srt_002(t *testing.T) {
	// Without the markdown namespace, there's no metadata
	meta, err := metadata.GetMetadata(context.Background(), strings.NewReader(srt), "application/x-subrip")
	if err != nil || len(meta) != 0 {
		t.Errorf("got %v, %v", meta, err)
	}
}
