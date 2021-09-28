package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	// Packages
	"github.com/djthorpe/go-media/pkg/media"

	// Namespace imports
	. "github.com/djthorpe/go-media"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type document struct {
	path  string
	info  fs.FileInfo
	meta  map[MediaKey]interface{}
	flags MediaFlag
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDocument(path string, info fs.FileInfo, media *media.MediaInput) (*document, error) {
	document := new(document)

	// Set up document
	document.path = path
	document.info = info
	document.meta = make(map[MediaKey]interface{})
	for _, k := range media.Metadata().Keys() {
		document.meta[k] = media.Metadata().Value(k)
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
	if name := d.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if path := d.Path(); path != "" {
		str += fmt.Sprintf(" path=%q", path)
	}
	if ext := d.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if modtime := d.ModTime(); !modtime.IsZero() {
		str += fmt.Sprint(" modtime=", modtime.Format(time.RFC3339))
	}
	if size := d.Size(); size > 0 {
		str += fmt.Sprint(" size=", size)
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
	if title := d.meta[MEDIA_KEY_TITLE]; title != nil {
		return title.(string)
	} else {
		return filepath.Base(d.path)
	}
}

func (d *document) Description() string {
	if desc := d.meta[MEDIA_KEY_DESCRIPTION]; desc != nil {
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

func (d *document) File() DocumentFile {
	return d
}

func (d *document) Meta() map[DocumentKey]interface{} {
	m := make(map[DocumentKey]interface{}, len(d.meta))
	for k, v := range d.meta {
		key := DocumentKey(k)
		m[key] = v
	}
	return m
}

func (d *document) HTML() []DocumentSection {
	return nil
}

func (d *document) Name() string {
	return d.info.Name()
}

func (d *document) Path() string {
	return filepath.Dir(d.path)
}

func (d *document) Ext() string {
	return filepath.Ext(d.path)
}

func (d *document) ModTime() time.Time {
	return d.info.ModTime()
}

func (d *document) Size() int64 {
	return d.info.Size()
}
