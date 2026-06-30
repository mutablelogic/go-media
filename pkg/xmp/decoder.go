package xmp

import (
	"encoding/xml"
	"io"
	"strings"

	media "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	nsRDF   = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	nsXML   = "http://www.w3.org/XML/1998/namespace"
	nsXMeta = "adobe:ns:meta/"

	maxFileSize     = 1 << 20 // 1 MiB
	maxNestingDepth = 64
)

// knownPrefixes maps well-known XMP namespace URIs to their conventional prefix.
var knownPrefixes = map[string]string{
	nsRDF:   "rdf",
	nsXMeta: "x",
	"http://purl.org/dc/elements/1.1/":               "dc",
	"http://ns.adobe.com/xap/1.0/":                   "xmp",
	"http://ns.adobe.com/xap/1.0/mm/":                "xmpMM",
	"http://ns.adobe.com/xap/1.0/rights/":            "xmpRights",
	"http://ns.adobe.com/xap/1.0/bj/":                "xmpBJ",
	"http://ns.adobe.com/xap/1.0/t/pg/":              "xmpTPg",
	"http://iptc.org/std/Iptc4xmpCore/1.0/xmlns/":    "Iptc4xmpCore",
	"http://ns.adobe.com/photoshop/1.0/":             "photoshop",
	"http://ns.adobe.com/exif/1.0/":                  "exif",
	"http://ns.adobe.com/exif/1.0/aux/":              "aux",
	"http://ns.adobe.com/tiff/1.0/":                  "tiff",
	"http://ns.adobe.com/camera-raw-settings/1.0/":   "crs",
	"http://ns.adobe.com/xap/1.0/sType/ResourceRef#": "stRef",
	"http://ns.adobe.com/xap/1.0/sType/Version#":     "stVer",
	"http://ns.adobe.com/xap/1.0/sType/Font#":        "stFnt",
	"http://ns.adobe.com/xap/1.0/sType/Dimensions#":  "stDim",
	"http://ns.adobe.com/xap/1.0/sType/ResourceEvent#": "stEvt",
}

////////////////////////////////////////////////////////////////////////////////
// TYPES

type xmpDecoder struct {
	d     *xml.Decoder
	ns    map[string]string // URI → prefix (built up during parsing)
	depth int               // current XML element nesting depth
}

// limitedReader returns an error (not io.EOF) when n bytes have been read,
// so that xml.Decoder propagates the violation rather than treating it as a
// clean end-of-file.
type limitedReader struct {
	r io.Reader
	n int64
}

func (lr *limitedReader) Read(p []byte) (int, error) {
	if lr.n <= 0 {
		return 0, media.ErrBadParameter.Withf("XMP document exceeds %d byte size limit", maxFileSize)
	}
	if int64(len(p)) > lr.n {
		p = p[:lr.n]
	}
	n, err := lr.r.Read(p)
	lr.n -= int64(n)
	return n, err
}

////////////////////////////////////////////////////////////////////////////////
// ENTRY POINT

func decode(r io.Reader) (*XMP, error) {
	d := &xmpDecoder{
		d:  xml.NewDecoder(&limitedReader{r: r, n: maxFileSize}),
		ns: make(map[string]string),
	}
	return d.decode()
}

func (d *xmpDecoder) decode() (*XMP, error) {
	for {
		tok, err := d.token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		d.collectNS(se.Attr)
		if se.Name.Space == nsRDF && se.Name.Local == "RDF" {
			return d.decodeRDF()
		}
		// x:xmpmeta or other wrappers — keep descending
	}
	return nil, media.ErrBadParameter.With("no XMP/RDF data found")
}

////////////////////////////////////////////////////////////////////////////////
// RDF

func (d *xmpDecoder) decodeRDF() (*XMP, error) {
	xmp := &XMP{}
	for {
		tok, err := d.token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			d.collectNS(t.Attr)
			if t.Name.Space == nsRDF && t.Name.Local == "Description" {
				if err := d.decodeDescription(xmp, t); err != nil {
					return nil, err
				}
			} else {
				if err := d.skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			return xmp, nil
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// rdf:Description

func (d *xmpDecoder) decodeDescription(xmp *XMP, start xml.StartElement) error {
	d.collectNS(start.Attr)

	// Inline attributes on rdf:Description are Simple properties
	for _, attr := range start.Attr {
		switch {
		case attr.Name.Space == nsRDF && attr.Name.Local == "about":
			xmp.about = attr.Value
		case attr.Name.Space == "xmlns", attr.Name.Space == nsXML, attr.Name.Space == "":
			// namespace declaration or plain attribute — skip
		default:
			xmp.items = append(xmp.items, &Item{
				ns:     attr.Name.Space,
				prefix: d.prefixFor(attr.Name.Space),
				name:   attr.Name.Local,
				kind:   Simple,
				value:  attr.Value,
			})
		}
	}

	// Child elements are properties
	for {
		tok, err := d.token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			item, err := d.decodeProperty(t)
			if err != nil {
				return err
			}
			if item != nil {
				xmp.items = append(xmp.items, item)
			}
		case xml.EndElement:
			return nil
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// Property elements

func (d *xmpDecoder) decodeProperty(start xml.StartElement) (*Item, error) {
	d.collectNS(start.Attr)
	prefix := d.prefixFor(start.Name.Space)

	// rdf:resource attribute → Simple URI value
	for _, attr := range start.Attr {
		if attr.Name.Space == nsRDF && attr.Name.Local == "resource" {
			if err := d.consumeEnd(start.Name); err != nil {
				return nil, err
			}
			return &Item{
				ns: start.Name.Space, prefix: prefix,
				name: start.Name.Local, kind: Simple, value: attr.Value,
			}, nil
		}
		// rdf:parseType="Resource" is a shorthand for an inline struct
		if attr.Name.Space == nsRDF && attr.Name.Local == "parseType" && attr.Value == "Resource" {
			fields, err := d.decodeStructFields(start)
			if err != nil {
				return nil, err
			}
			return &Item{
				ns: start.Name.Space, prefix: prefix,
				name: start.Name.Local, kind: Struct, items: fields,
			}, nil
		}
	}

	for {
		tok, err := d.token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s == "" {
				continue
			}
			// Text content → Simple
			if err := d.consumeEnd(start.Name); err != nil {
				return nil, err
			}
			return &Item{
				ns: start.Name.Space, prefix: prefix,
				name: start.Name.Local, kind: Simple,
				lang: xmlLang(start.Attr), value: s,
			}, nil

		case xml.StartElement:
			if t.Name.Space == nsRDF {
				var (
					children []*Item
					kind     Kind
					decErr   error
				)
				switch t.Name.Local {
				case "Bag":
					kind, children, decErr = Bag, nil, nil
					children, decErr = d.decodeListItems()
				case "Seq":
					kind, children, decErr = Seq, nil, nil
					children, decErr = d.decodeListItems()
				case "Alt":
					kind, children, decErr = Alt, nil, nil
					children, decErr = d.decodeListItems()
				case "Description":
					kind = Struct
					children, decErr = d.decodeStructFields(t)
				default:
					if err := d.skip(); err != nil {
						return nil, err
					}
					continue
				}
				if decErr != nil {
					return nil, decErr
				}
				if err := d.consumeEnd(start.Name); err != nil {
					return nil, err
				}
				return &Item{
					ns: start.Name.Space, prefix: prefix,
					name: start.Name.Local, kind: kind, items: children,
				}, nil
			}
			// Unknown nested element — skip and keep looking
			if err := d.skip(); err != nil {
				return nil, err
			}

		case xml.EndElement:
			// Self-closing or empty element → empty Simple
			return &Item{
				ns: start.Name.Space, prefix: prefix,
				name: start.Name.Local, kind: Simple, value: "",
			}, nil
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// rdf:Bag / rdf:Seq / rdf:Alt  (reads rdf:li items until container end)

func (d *xmpDecoder) decodeListItems() ([]*Item, error) {
	var items []*Item
	for {
		tok, err := d.token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Space == nsRDF && t.Name.Local == "li" {
				item, err := d.decodeLi(t)
				if err != nil {
					return nil, err
				}
				if item != nil {
					items = append(items, item)
				}
			} else {
				if err := d.skip(); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			return items, nil
		}
	}
}

func (d *xmpDecoder) decodeLi(start xml.StartElement) (*Item, error) {
	lang := xmlLang(start.Attr)

	// rdf:resource / rdf:parseType="Resource" on rdf:li
	for _, attr := range start.Attr {
		if attr.Name.Space != nsRDF {
			continue
		}
		if attr.Name.Local == "resource" {
			if err := d.consumeEnd(start.Name); err != nil {
				return nil, err
			}
			return &Item{ns: nsRDF, prefix: "rdf", name: "li", kind: Simple, lang: lang, value: attr.Value}, nil
		}
		if attr.Name.Local == "parseType" && attr.Value == "Resource" {
			// Children are inline struct fields, no wrapping rdf:Description
			fields, err := d.decodeStructFields(start)
			if err != nil {
				return nil, err
			}
			return &Item{ns: nsRDF, prefix: "rdf", name: "li", kind: Struct, lang: lang, items: fields}, nil
		}
	}

	var buf strings.Builder
	for {
		tok, err := d.token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write(t)
		case xml.StartElement:
			// Struct inside an li (rdf:Description)
			if t.Name.Space == nsRDF && t.Name.Local == "Description" {
				fields, err := d.decodeStructFields(t)
				if err != nil {
					return nil, err
				}
				if err := d.consumeEnd(start.Name); err != nil {
					return nil, err
				}
				return &Item{ns: nsRDF, prefix: "rdf", name: "li", kind: Struct, lang: lang, items: fields}, nil
			}
			if err := d.skip(); err != nil {
				return nil, err
			}
		case xml.EndElement:
			return &Item{
				ns: nsRDF, prefix: "rdf", name: "li", kind: Simple,
				lang: lang, value: strings.TrimSpace(buf.String()),
			}, nil
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// rdf:Description inside a property (struct value)

func (d *xmpDecoder) decodeStructFields(start xml.StartElement) ([]*Item, error) {
	d.collectNS(start.Attr)
	var fields []*Item

	// Inline attributes are struct fields
	for _, attr := range start.Attr {
		switch {
		case attr.Name.Space == "xmlns", attr.Name.Space == nsRDF,
			attr.Name.Space == nsXML, attr.Name.Space == "":
			// skip namespace declarations and rdf: control attributes
		default:
			fields = append(fields, &Item{
				ns:     attr.Name.Space,
				prefix: d.prefixFor(attr.Name.Space),
				name:   attr.Name.Local,
				kind:   Simple,
				value:  attr.Value,
			})
		}
	}

	// Child elements are struct fields
	for {
		tok, err := d.token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			item, err := d.decodeProperty(t)
			if err != nil {
				return nil, err
			}
			if item != nil {
				fields = append(fields, item)
			}
		case xml.EndElement:
			return fields, nil
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// HELPERS

// token wraps d.d.Token(), tracking nesting depth and enforcing the limit.
func (d *xmpDecoder) token() (xml.Token, error) {
	tok, err := d.d.Token()
	if err != nil {
		return nil, err
	}
	switch tok.(type) {
	case xml.StartElement:
		d.depth++
		if d.depth > maxNestingDepth {
			return nil, media.ErrBadParameter.Withf("XMP nesting depth exceeds %d", maxNestingDepth)
		}
	case xml.EndElement:
		d.depth--
	}
	return tok, nil
}

// skip consumes the children and closing element of a StartElement that has
// already been read. Uses token() so that d.depth stays accurate.
func (d *xmpDecoder) skip() error {
	depth := 1
	for {
		tok, err := d.token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
			if depth == 0 {
				return nil
			}
		}
	}
}

// collectNS extracts xmlns: declarations from attrs into d.ns.
func (d *xmpDecoder) collectNS(attrs []xml.Attr) {
	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" && attr.Name.Local != "" {
			d.ns[attr.Value] = attr.Name.Local
		}
	}
}

// prefixFor returns the preferred prefix for a namespace URI.
func (d *xmpDecoder) prefixFor(uri string) string {
	if p, ok := d.ns[uri]; ok {
		return p
	}
	if p, ok := knownPrefixes[uri]; ok {
		return p
	}
	return ""
}

// consumeEnd discards tokens until the matching end element for name is found.
// Uses token() to keep d.depth current.
func (d *xmpDecoder) consumeEnd(name xml.Name) error {
	sameNameDepth := 0
	for {
		tok, err := d.token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name == name {
				sameNameDepth++
			}
		case xml.EndElement:
			if t.Name == name {
				if sameNameDepth == 0 {
					return nil
				}
				sameNameDepth--
			}
		}
	}
}

// xmlLang extracts the xml:lang attribute value from attrs.
func xmlLang(attrs []xml.Attr) string {
	for _, attr := range attrs {
		if attr.Name.Space == nsXML && attr.Name.Local == "lang" {
			return attr.Value
		}
	}
	return ""
}
