package metadata_test

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/metadata"
)

// markdownDocument is converted a line at a time by the markdown handler,
// with characters of two, three and four bytes
const markdownDocument = "# Café\n\nÜber naïve façades — 日本語 🎬 and more.\n\n- one\n- two\n"

// markdownLines counts the lines the handler has written
var markdownLines int

func init() {
	// The handler converts the document a line at a time, and stops when the
	// writer has the part. A document which is "fail" is an error.
	AddHandler(regexp.MustCompile("^x-test/markdown$"), "markdown", func(_ context.Context, r io.Reader, o *Opts) ([]gomedia.Metadata, error) {
		if !o.HasNamespace(gomedia.MetaMarkdown) {
			return nil, nil
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		} else if string(data) == "fail" {
			return nil, errors.New("conversion failed")
		}
		w := NewMarkdownWriter(o)
		for _, line := range strings.SplitAfter(string(data), "\n") {
			markdownLines++
			if _, err = w.WriteString(line); err != nil {
				break
			}
		}
		return w.Metadata(err)
	}, gomedia.MetaMarkdown)
}

// markdownPart returns the body, offset and next of the document's Markdown
// in the range
func markdownPart(t *testing.T, document string, opts ...Option) (string, uint64, *uint64, error) {
	t.Helper()
	opts = append(opts, WithNamespace(gomedia.MetaMarkdown))
	meta, err := GetMetadata(context.Background(), strings.NewReader(document), "x-test/markdown", opts...)
	if err != nil {
		return "", 0, nil, err
	}
	var body string
	var offset uint64
	var next *uint64
	for _, entry := range meta {
		switch entry.Key() {
		case "markdown:body":
			body = entry.Any().(string)
		case "markdown:offset":
			offset = entry.Any().(uint64)
		case "markdown:next":
			n := entry.Any().(uint64)
			next = &n
		default:
			t.Errorf("unexpected key %q", entry.Key())
		}
	}
	return body, offset, next, nil
}

func Test_markdown_000(t *testing.T) {
	// Without a range, the body is all of the Markdown
	body, offset, next, err := markdownPart(t, markdownDocument)
	if err != nil {
		t.Fatal(err)
	} else if body != markdownDocument || offset != 0 || next != nil {
		t.Errorf("got %q from %d, next %v", body, offset, next)
	}

	// Without the namespace, there's none
	meta, err := GetMetadata(context.Background(), strings.NewReader(markdownDocument), "x-test/markdown")
	if err != nil || len(meta) != 0 {
		t.Errorf("without the namespace: %v, %v", meta, err)
	}
}

func Test_markdown_001(t *testing.T) {
	// Reading a part at a time, from each next, reads all of the Markdown,
	// in whole characters, for every limit
	for limit := uint64(utf8.UTFMax); limit <= uint64(len(markdownDocument))+1; limit++ {
		var result strings.Builder
		offset := uint64(0)
		for parts := 0; ; parts++ {
			if parts > len(markdownDocument) {
				t.Fatalf("limit %d: too many parts", limit)
			}
			body, start, next, err := markdownPart(t, markdownDocument, WithRange(offset, limit))
			if err != nil {
				t.Fatalf("limit %d, offset %d: %v", limit, offset, err)
			} else if start != offset {
				t.Fatalf("limit %d: part starts at %d, not %d", limit, start, offset)
			} else if !utf8.ValidString(body) || uint64(len(body)) > limit {
				t.Fatalf("limit %d, offset %d: part %q", limit, offset, body)
			}
			result.WriteString(body)
			if next == nil {
				break
			} else if *next != offset+uint64(len(body)) || *next <= offset {
				t.Fatalf("limit %d, offset %d: next is %d", limit, offset, *next)
			}
			offset = *next
		}
		if result.String() != markdownDocument {
			t.Errorf("limit %d: read %q", limit, result.String())
		}
	}
}

func Test_markdown_002(t *testing.T) {
	// The handler stops once the writer has the part
	markdownLines = 0
	body, _, next, err := markdownPart(t, markdownDocument, WithRange(0, 4))
	if err != nil {
		t.Fatal(err)
	} else if body != "# Ca" || next == nil || *next != 4 {
		t.Errorf("got %q, next %v", body, next)
	} else if markdownLines != 1 {
		t.Errorf("the handler wrote %d lines, not 1", markdownLines)
	}
}

func Test_markdown_003(t *testing.T) {
	// An offset within a character starts at the next character
	offset := uint64(strings.Index(markdownDocument, "é") + 1)
	body, start, _, err := markdownPart(t, markdownDocument, WithRange(offset, 0))
	if err != nil {
		t.Fatal(err)
	} else if start != offset+1 || body != markdownDocument[offset+1:] {
		t.Errorf("got %q from %d", body, start)
	}

	// The end of the Markdown has an empty body, and beyond it is an error
	end := uint64(len(markdownDocument))
	if body, _, next, err := markdownPart(t, markdownDocument, WithRange(end, 0)); err != nil || body != "" || next != nil {
		t.Errorf("at the end: %q, %v, %v", body, next, err)
	}
	if _, _, _, err := markdownPart(t, markdownDocument, WithRange(end+1, 0)); !errors.Is(err, gomedia.ErrBadParameter) {
		t.Errorf("beyond the end: %v", err)
	}
}

func Test_markdown_004(t *testing.T) {
	// A handler's error is returned
	if _, _, _, err := markdownPart(t, "fail", WithRange(0, 16)); err == nil || err.Error() != "conversion failed" {
		t.Errorf("got %v", err)
	}

	// A limit holds a whole character
	if _, _, _, err := markdownPart(t, markdownDocument, WithRange(0, utf8.UTFMax-1)); !errors.Is(err, gomedia.ErrBadParameter) {
		t.Errorf("got %v", err)
	}
}
