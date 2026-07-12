package heif_test

import (
	"testing"

	"github.com/mutablelogic/go-media/pkg/heif"
)

func Test_heif_metadata_000(t *testing.T) {
	h, err := heif.Open(testHEIF)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	meta := h.Metadata()
	if len(meta) == 0 {
		t.Skip("no metadata blocks in fixture")
	}

	t.Logf("metadata count=%d", len(meta))
	for i, m := range meta {
		t.Logf("%03d key=%q value=%q bytes=%d any=%T(%v)", i, m.Key(), m.Value(), len(m.Bytes()), m.Any(), m.Any())
	}

	if doc := h.XMP(); doc != nil {
		t.Logf("xmp document items=%d", len(doc.Items()))
		t.Log(doc.String())
	}
}
