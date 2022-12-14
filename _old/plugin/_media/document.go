package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"strings"

	// Packages
	"github.com/mutablelogic/go-media/pkg/media"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type document struct {
	name  string
	meta  map[DocumentKey]interface{}
	flags MediaFlag
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(name string, media *media.MediaInput, meta map[DocumentKey]interface{}) (*document, error) {
	document := new(document)

	// Set up document
	document.name = name
	document.meta = make(map[DocumentKey]interface{})
	for k, v := range meta {
		document.meta[k] = v
	}
	for _, k := range media.Metadata().Keys() {
		key := DocumentKey(k)
		document.meta[key] = media.Metadata().Value(k)
	}
	document.flags = media.Flags() &^ (MEDIA_FLAG_ENCODER | MEDIA_FLAG_DECODER)

	// Return success
	return document, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d *document) String() string {
	str := "<document"
	if title := d.Title(); title != "" {
		str += fmt.Sprintf(" title=%q", title)
	}
	if desc := d.Description(); desc != "" {
		str += fmt.Sprintf(" description=%q", desc)
	}
	if shortform := d.Shortform(); shortform != "" {
		str += fmt.Sprintf(" shortform=%q", shortform)
	}
	if tags := d.Tags(); len(tags) > 0 {
		str += fmt.Sprintf(" tags=%q", tags)
	}
	for k, v := range d.meta {
		if _, ok := v.(string); ok {
			str += fmt.Sprintf(" %s=%q", k, v)
		} else {
			str += fmt.Sprintf(" %s=%v", k, v)
		}
	}
	return str + ">"
}

func (d *document) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString("{")
	if err := d.marshalKV(buf, "title", d.Title(), ","); err != nil {
		return nil, err
	}
	if err := d.marshalKV(buf, "description", d.Description(), ","); err != nil {
		return nil, err
	}
	if err := d.marshalKV(buf, "shortform", d.Shortform(), ","); err != nil {
		return nil, err
	}
	if err := d.marshalKV(buf, "tags", d.Tags(), ","); err != nil {
		return nil, err
	}
	if err := d.marshalKV(buf, "metadata", d.meta, "}"); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *document) marshalKV(w io.Writer, key string, value interface{}, suffix string) error {
	if k, err := json.Marshal(key); err != nil {
		return err
	} else if v, err := json.Marshal(value); err != nil {
		return err
	} else if _, err := w.Write(k); err != nil {
		return err
	} else if _, err := w.Write([]byte(": ")); err != nil {
		return err
	} else if _, err := w.Write(v); err != nil {
		return err
	} else if _, err := w.Write([]byte(suffix)); err != nil {
		return err
	} else {
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *document) Title() string {
	key := DocumentKey(MEDIA_KEY_TITLE)
	if title := d.meta[key]; title != nil {
		return title.(string)
	} else {
		return d.name
	}
}

func (d *document) Description() string {
	key := DocumentKey(MEDIA_KEY_DESCRIPTION)
	if desc := d.meta[key]; desc != nil {
		return desc.(string)
	} else {
		return ""
	}
}

func (d *document) Shortform() template.HTML {
	return ""
}

func (d *document) Tags() []string {
	tags := []string{}
	for f := MEDIA_FLAG_MIN; f <= MEDIA_FLAG_MAX; f <<= 1 {
		if d.flags.Is(f) {
			tag := strings.TrimPrefix(f.FlagString(), "MEDIA_FLAG_")
			tags = append(tags, strings.ToLower(tag))
		}
	}
	return tags
}

func (d *document) Meta() map[DocumentKey]interface{} {
	return d.meta
}

func (d *document) HTML() []DocumentSection {
	return nil
}
