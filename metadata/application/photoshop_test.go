package application

import (
	"bytes"
	"image"
	"io"
	"testing"

	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	"github.com/mutablelogic/go-media/pkg/xmp"
	psd "github.com/oov/psd"
)

type namedReader struct {
	io.Reader
	name string
}

func (r namedReader) Name() string { return r.name }

func metadataMap(items []gomedia.Metadata) map[string]gomedia.Metadata {
	out := make(map[string]gomedia.Metadata, len(items))
	for _, item := range items {
		out[item.Key()] = item
	}
	return out
}

func TestPhotoshopMetadataFromConfig(t *testing.T) {
	var x bytes.Buffer
	doc := xmp.New()
	doc.Add(
		xmp.NewItem("http://ns.adobe.com/photoshop/1.0/", "photoshop", "DateCreated", "2024-01-15"),
		xmp.NewItem("http://ns.adobe.com/xap/1.0/", "xmp", "CreateDate", "2024-01-16"),
	)
	if err := doc.Write(&x); err != nil {
		t.Fatalf("write xmp: %v", err)
	}

	items, err := photoshopMetadata(psd.Config{
		Version:   2,
		Rect:      image.Rect(0, 0, 640, 480),
		Channels:  4,
		Depth:     16,
		ColorMode: 3,
		Res: map[int]psd.ImageResource{
			photoshopXMPResourceID: {Data: x.Bytes()},
		},
	}, "")
	if err != nil {
		t.Fatalf("photoshopMetadata: %v", err)
	}

	got := metadataMap(items)
	checks := map[string]string{
		"photoshop:Format":      "PSB",
		"photoshop:Version":     "2",
		"photoshop:Width":       "640",
		"photoshop:Height":      "480",
		"photoshop:Channels":    "4",
		"photoshop:Depth":       "16",
		"photoshop:ColorMode":   "3",
		"photoshop:DateCreated": "2024-01-15",
		"xmp:CreateDate":        "2024-01-16",
	}
	for key, want := range checks {
		item, ok := got[key]
		if !ok {
			t.Fatalf("missing metadata key %q", key)
		}
		if got := item.Value(); got != want {
			t.Fatalf("%s: want %q got %q", key, want, got)
		}
	}

	filtered, err := photoshopMetadata(psd.Config{
		Version:   2,
		Rect:      image.Rect(0, 0, 640, 480),
		Channels:  4,
		Depth:     16,
		ColorMode: 3,
		Res: map[int]psd.ImageResource{
			photoshopXMPResourceID: {Data: x.Bytes()},
		},
	}, "xmp:")
	if err != nil {
		t.Fatalf("photoshopMetadata(filter): %v", err)
	}
	for _, item := range filtered {
		if item.Key() != "xmp:CreateDate" {
			t.Fatalf("unexpected filtered key %q", item.Key())
		}
	}
}

func TestPhotoshopContentTypeByExtension(t *testing.T) {
	contentType, _, err := metadata.ContentType(namedReader{Reader: bytes.NewReader([]byte{0x00}), name: "sample.psd"})
	if err != nil {
		t.Fatalf("ContentType: %v", err)
	}
	if contentType != "application/vnd.adobe.photoshop" {
		t.Fatalf("want application/vnd.adobe.photoshop, got %q", contentType)
	}
}
