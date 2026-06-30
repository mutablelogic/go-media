package xmp

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	xpacketBegin = "<?xpacket begin=\"\xef\xbb\xbf\" id=\"W5M0MpCehiHzreSzNTczkc9d\"?>\n"
	xpacketEnd   = "\n<?xpacket end=\"w\"?>\n"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type xmpEncoder struct {
	w *bufio.Writer
}

////////////////////////////////////////////////////////////////////////////////
// ENTRY POINT

func encode(w io.Writer, x *XMP) error {
	enc := &xmpEncoder{w: bufio.NewWriter(w)}
	if err := enc.encode(x); err != nil {
		return err
	}
	return enc.w.Flush()
}

func (e *xmpEncoder) encode(x *XMP) error {
	// Collect namespace declarations needed for all items
	nsMap := make(map[string]string) // URI → prefix
	for _, it := range x.items {
		collectNS(it, nsMap)
	}

	e.w.WriteString(xpacketBegin)
	e.w.WriteString("<x:xmpmeta xmlns:x=\"adobe:ns:meta/\">\n")
	e.w.WriteString("  <rdf:RDF xmlns:rdf=\"http://www.w3.org/1999/02/22-rdf-syntax-ns#\">\n")

	// rdf:Description opener with rdf:about and namespace declarations
	fmt.Fprintf(e.w, "    <rdf:Description rdf:about=%q", x.about)
	uris := make([]string, 0, len(nsMap))
	for uri := range nsMap {
		uris = append(uris, uri)
	}
	sort.Strings(uris)
	for _, uri := range uris {
		fmt.Fprintf(e.w, "\n        xmlns:%s=%q", nsMap[uri], uri)
	}

	if len(x.items) == 0 {
		e.w.WriteString("/>\n")
	} else {
		e.w.WriteString(">\n")
		for _, it := range x.items {
			if err := e.writeItem(it, "      "); err != nil {
				return err
			}
		}
		e.w.WriteString("    </rdf:Description>\n")
	}

	e.w.WriteString("  </rdf:RDF>\n")
	e.w.WriteString("</x:xmpmeta>")
	e.w.WriteString(xpacketEnd)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ITEM WRITING

func (e *xmpEncoder) writeItem(it *Item, indent string) error {
	tag := itemTag(it)

	switch it.kind {
	case Simple:
		if it.lang != "" {
			fmt.Fprintf(e.w, "%s<%s xml:lang=%q>%s</%s>\n",
				indent, tag, it.lang, escapeXML(it.value), tag)
		} else {
			fmt.Fprintf(e.w, "%s<%s>%s</%s>\n",
				indent, tag, escapeXML(it.value), tag)
		}

	case Bag:
		fmt.Fprintf(e.w, "%s<%s>\n%s  <rdf:Bag>\n", indent, tag, indent)
		for _, child := range it.items {
			e.writeLi(child, indent+"    ")
		}
		fmt.Fprintf(e.w, "%s  </rdf:Bag>\n%s</%s>\n", indent, indent, tag)

	case Seq:
		fmt.Fprintf(e.w, "%s<%s>\n%s  <rdf:Seq>\n", indent, tag, indent)
		for _, child := range it.items {
			e.writeLi(child, indent+"    ")
		}
		fmt.Fprintf(e.w, "%s  </rdf:Seq>\n%s</%s>\n", indent, indent, tag)

	case Alt:
		fmt.Fprintf(e.w, "%s<%s>\n%s  <rdf:Alt>\n", indent, tag, indent)
		for _, child := range it.items {
			e.writeLi(child, indent+"    ")
		}
		fmt.Fprintf(e.w, "%s  </rdf:Alt>\n%s</%s>\n", indent, indent, tag)

	case Struct:
		fmt.Fprintf(e.w, "%s<%s>\n%s  <rdf:Description>\n", indent, tag, indent)
		for _, field := range it.items {
			if err := e.writeItem(field, indent+"    "); err != nil {
				return err
			}
		}
		fmt.Fprintf(e.w, "%s  </rdf:Description>\n%s</%s>\n", indent, indent, tag)
	}
	return nil
}

func (e *xmpEncoder) writeLi(it *Item, indent string) {
	if it.lang != "" {
		fmt.Fprintf(e.w, "%s<rdf:li xml:lang=%q>%s</rdf:li>\n",
			indent, it.lang, escapeXML(it.value))
	} else {
		fmt.Fprintf(e.w, "%s<rdf:li>%s</rdf:li>\n",
			indent, escapeXML(it.value))
	}
}

////////////////////////////////////////////////////////////////////////////////
// HELPERS

// itemTag returns "prefix:name" or just "name" if no prefix.
func itemTag(it *Item) string {
	if it.prefix == "" {
		return it.name
	}
	return it.prefix + ":" + it.name
}

// collectNS recursively gathers namespace URI→prefix mappings from item and
// all its descendants, skipping the RDF namespace (declared at the RDF level).
func collectNS(it *Item, ns map[string]string) {
	if it.ns != "" && it.ns != nsRDF && it.prefix != "" {
		ns[it.ns] = it.prefix
	}
	for _, child := range it.items {
		collectNS(child, ns)
	}
}

// escapeXML returns s with XML special characters escaped.
func escapeXML(s string) string {
	var buf strings.Builder
	xml.EscapeText(&buf, []byte(s)) //nolint:errcheck // strings.Builder never errors
	return buf.String()
}
