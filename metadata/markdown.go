package metadata

import (
	"errors"
	"fmt"
	"image"
	"unicode/utf8"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// MarkdownWriter is written the content of a document as Markdown, by a
// handler of the "markdown" namespace, and keeps the part of it in the range
// set via WithRange. Once it has the part, a write returns an error, so the
// handler can stop, rather than converting the rest of the document. A
// handler returns the part with Metadata, which also takes the handler's
// error.
type MarkdownWriter struct {
	offset, limit uint64
	pos           uint64
	buf           []byte
	more          bool
}

// markdown is an entry of the "markdown" namespace
type markdown struct {
	key   string
	value any
}

var _ gomedia.Metadata = markdown{}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// errMarkdownFull is returned by a write once the writer has the part of the
// Markdown in its range
var errMarkdownFull = errors.New("the markdown's range is full")

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewMarkdownWriter returns a writer which keeps the part of the Markdown in
// the range set via WithRange, or all of it
func NewMarkdownWriter(o *Opts) *MarkdownWriter {
	return &MarkdownWriter{offset: o.offset, limit: o.limit}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write keeps the part of the data in the range. It returns an error once it
// has the part, and there's more of the Markdown, so the handler can stop.
func (w *MarkdownWriter) Write(data []byte) (int, error) {
	if w.more {
		return 0, errMarkdownFull
	}
	n := 0

	// Skip the data before the offset
	if w.pos < w.offset {
		skip := int(min(uint64(len(data)), w.offset-w.pos))
		w.pos += uint64(skip)
		data, n = data[skip:], skip
	}

	// Keep the data up to the limit, and one more byte, so there's more.
	// What's kept is at most the data, as the buffer is within the limit.
	if w.limit > 0 && uint64(len(w.buf))+uint64(len(data)) > w.limit {
		keep := int(min(w.limit-uint64(len(w.buf))+1, uint64(len(data))))
		w.buf = append(w.buf, data[:keep]...)
		w.pos += uint64(keep)
		w.more = true
		return n + keep, errMarkdownFull
	}
	w.buf = append(w.buf, data...)
	w.pos += uint64(len(data))
	return n + len(data), nil
}

// WriteString keeps the part of the string in the range, as Write does
func (w *MarkdownWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// Metadata returns the part of the Markdown, and the handler's error, which
// is nil when it stopped as the writer had the part. The part is
// markdown:body; markdown:offset is the byte it starts at, when the range
// has an offset, which is moved forward to the start of a character; and
// markdown:next is the offset of the next part, when there's more. Returns
// an error when the offset is beyond the end of the Markdown.
func (w *MarkdownWriter) Metadata(err error) ([]gomedia.Metadata, error) {
	if err != nil && !errors.Is(err, errMarkdownFull) {
		return nil, err
	} else if w.pos < w.offset {
		return nil, gomedia.ErrBadParameter.Withf("offset %d is beyond the end of the markdown, at %d", w.offset, w.pos)
	}

	// The part starts and ends with whole characters
	data := w.buf
	start := 0
	if w.offset > 0 {
		for start < len(data) && start < utf8.UTFMax && !utf8.RuneStart(data[start]) {
			start++
		}
	}
	end := len(data)
	if w.more {
		end = start + validEnd(data[start:w.limit])
	}

	result := []gomedia.Metadata{
		markdown{key: gomedia.MetaMarkdown + ":body", value: string(data[start:end])},
	}
	if w.offset > 0 {
		result = append(result, markdown{key: gomedia.MetaMarkdown + ":offset", value: w.offset + uint64(start)})
	}
	if w.more {
		result = append(result, markdown{key: gomedia.MetaMarkdown + ":next", value: w.offset + uint64(end)})
	}
	return result, nil
}

func (m markdown) Key() string        { return m.key }
func (m markdown) Value() string      { return fmt.Sprint(m.value) }
func (m markdown) Bytes() []byte      { return nil }
func (m markdown) Image() image.Image { return nil }
func (m markdown) Any() any           { return m.value }

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// validEnd returns the end of the data without a character which is cut off
// at its end, which is that character's start
func validEnd(data []byte) int {
	for i := len(data) - 1; i >= 0 && i >= len(data)-utf8.UTFMax; i-- {
		if utf8.RuneStart(data[i]) {
			if utf8.FullRune(data[i:]) {
				return len(data)
			}
			return i
		}
	}
	return len(data)
}
