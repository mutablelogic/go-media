package image

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	htmlindex "golang.org/x/text/encoding/htmlindex"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	nsSVG = "http://www.w3.org/2000/svg"
	nsDC  = "http://purl.org/dc/elements/1.1/"
	nsRDF = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
)

// maxSVGSize is the most of an SVG image which is read. The metadata is
// usually at the start, before the drawing, which can be large.
const maxSVGSize = 16 << 20

// svgDateLayouts are the layouts of a dc:date, which is a W3C date and time
var svgDateLayouts = []string{time.RFC3339Nano, "2006-01-02T15:04Z07:00", "2006-01-02T15:04", "2006-01-02", "2006-01", "2006"}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// The generic image handler has no decoder for SVG, so this handler
	// returns its size as well as its title, description and Dublin Core
	// metadata
	metadata.AddHandler(regexp.MustCompile(`^image/svg\+xml$`), "svg", func(_ context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		entries, err := svgMetadata(r)
		if err != nil {
			return nil, err
		}
		return metadata.FilterMetadata(entries, o), nil
	}, "image", "dc")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// svgMetadata returns the size of an SVG image, its title and description,
// from its title and desc elements, and the Dublin Core metadata in its
// metadata element, as Inkscape writes. An image which isn't well-formed
// returns the metadata read before the error, as other handlers may still
// read it; only an error reading it is returned.
func svgMetadata(r io.Reader) (map[string]gomedia.Metadata, error) {
	d := xml.NewDecoder(io.LimitReader(r, maxSVGSize))
	d.Strict = false
	d.CharsetReader = func(label string, input io.Reader) (io.Reader, error) {
		enc, err := htmlindex.Get(label)
		if err != nil {
			return nil, err
		}
		return enc.NewDecoder().Reader(input), nil
	}

	entries := make(map[string]gomedia.Metadata)
	set := func(key string, value any) {
		if _, exists := entries[key]; !exists {
			entries[key] = &imageMetadata{key, value}
		}
	}

	// The Dublin Core in the metadata element, which is used when the
	// title and desc elements aren't there
	dc := make(map[string][]string)
	var order []string

	depth := 0
read:
	for {
		token, err := d.Token()
		if svgEnd(err) {
			break
		} else if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			switch {
			case depth == 1:
				// The root element, which is the image's size
				if !isSVG(t.Name, "svg") {
					return entries, nil
				}
				set("image:format", "svg")
				if width, height, ok := svgSize(t.Attr); ok {
					set("image:width", width)
					set("image:height", height)
				}
			case depth == 2 && isSVG(t.Name, "title"):
				if text := svgText(d); text != "" {
					set("dc:title", text)
				}
				depth--
			case depth == 2 && isSVG(t.Name, "desc"):
				if text := svgText(d); text != "" {
					set("dc:description", text)
				}
				depth--
			case t.Name.Space == nsDC:
				// A Dublin Core element, wherever it is in the metadata,
				// which can be a list, or a name in an agent
				key := "dc:" + t.Name.Local
				values := svgValues(d)
				if len(values) > 0 {
					if _, exists := dc[key]; !exists {
						order = append(order, key)
					}
					dc[key] = append(dc[key], values...)
				}
				depth--
			case depth == 2 && !isSVG(t.Name, "metadata"):
				// Skip the drawing, which has no metadata. When it's cut off,
				// the metadata read before it is still returned.
				if err := d.Skip(); svgEnd(err) {
					break read
				} else if err != nil {
					return nil, err
				}
				depth--
			}
		case xml.EndElement:
			depth--
		}
	}

	// Add the Dublin Core metadata, where the title and description are a
	// single value, the subject is a list of keywords, the date is a time
	// when it can be parsed, and other elements are a list when there's
	// more than one
	for _, key := range order {
		values := dc[key]
		switch key {
		case "dc:title", "dc:description":
			set(key, values[0])
		case "dc:subject":
			set(key, values)
		case "dc:date":
			// A date which can't be parsed, such as a range, is kept as
			// its text
			if date, ok := svgDate(values[0]); ok {
				entries[key] = timeMetadata{key: key, t: date}
			} else {
				set(key, values[0])
			}
		default:
			if len(values) == 1 {
				set(key, values[0])
			} else {
				set(key, values)
			}
		}
	}

	return entries, nil
}

// svgEnd returns true when an error is the end of what can be read: the end
// of the image, the most which is read, or where it isn't well-formed
func svgEnd(err error) bool {
	var syntaxErr *xml.SyntaxError
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &syntaxErr)
}

// isSVG returns true when the name is the local name in the SVG namespace,
// or without a namespace, which some images don't declare
func isSVG(name xml.Name, local string) bool {
	return name.Local == local && (name.Space == nsSVG || name.Space == "")
}

// svgText returns the text in an element, including in the elements in it,
// with its whitespace collapsed, and reads to the end of the element
func svgText(d *xml.Decoder) string {
	var b strings.Builder
	for depth := 1; depth > 0; {
		token, err := d.Token()
		if err != nil {
			break
		}
		// Elements separate words, but pieces of text in an element, such
		// as before and in a CDATA section, don't
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			b.WriteByte(' ')
		case xml.EndElement:
			depth--
			b.WriteByte(' ')
		case xml.CharData:
			b.Write(t)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// svgValues returns the values of a Dublin Core element: the text of each
// rdf:li in a list, such as the keywords of dc:subject, or else the text in
// the element, such as the name in Inkscape's cc:Agent for dc:creator. It
// reads to the end of the element.
func svgValues(d *xml.Decoder) []string {
	var values []string
	var b strings.Builder
	for depth := 1; depth > 0; {
		token, err := d.Token()
		if err != nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Space == nsRDF && t.Name.Local == "li" {
				if text := svgText(d); text != "" {
					values = append(values, text)
				}
			} else {
				depth++
				b.WriteByte(' ')
			}
		case xml.EndElement:
			depth--
			b.WriteByte(' ')
		case xml.CharData:
			b.Write(t)
		}
	}
	if len(values) > 0 {
		return values
	}
	if text := strings.Join(strings.Fields(b.String()), " "); text != "" {
		return []string{text}
	}
	return nil
}

// svgSize returns the width and height of an image, in pixels, from its
// width and height attributes, when they're in pixels or have no unit, or
// else from its viewBox
func svgSize(attrs []xml.Attr) (int, int, bool) {
	var width, height, viewBox string
	for _, attr := range attrs {
		if attr.Name.Space != "" {
			continue
		}
		switch attr.Name.Local {
		case "width":
			width = attr.Value
		case "height":
			height = attr.Value
		case "viewBox":
			viewBox = attr.Value
		}
	}
	if w, ok := svgPixels(width); ok {
		if h, ok := svgPixels(height); ok {
			return w, h, true
		}
	}

	// The viewBox is the minimum x and y, the width and the height,
	// separated by whitespace or commas
	fields := strings.FieldsFunc(viewBox, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	if len(fields) == 4 {
		if w, ok := svgPixels(fields[2]); ok {
			if h, ok := svgPixels(fields[3]); ok {
				return w, h, true
			}
		}
	}
	return 0, 0, false
}

// svgPixels returns a length in whole pixels, when it's in pixels or has no
// unit, and is more than zero
func svgPixels(length string) (int, bool) {
	length = strings.TrimSuffix(strings.TrimSpace(length), "px")
	value, err := strconv.ParseFloat(length, 64)
	if err != nil || math.IsNaN(value) || value <= 0 || math.IsInf(value, 0) || value > math.MaxInt32 {
		return 0, false
	}
	return int(math.Round(value)), true
}

// svgDate returns the time of a dc:date, which is a W3C date and time
func svgDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	for _, layout := range svgDateLayouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
