package exif_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mutablelogic/go-media/pkg/exif"
	libheif "github.com/mutablelogic/go-media/sys/libheif"
)

const (
	testJPEG          = "../../etc/test/sample.jpg"
	testJPEGMakerNote = "../../etc/test/canon_makernote_variant_1.jpg"
)

////////////////////////////////////////////////////////////////////////////////
// OPEN / PARSE / READ

func Test_exif_000(t *testing.T) {
	e, err := exif.Open(testJPEG)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	t.Log(e)
}

func Test_exif_001(t *testing.T) {
	// Non-existent file should return an error.
	_, err := exif.Open("/does/not/exist.jpg")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	t.Log(err)
}

func Test_exif_002(t *testing.T) {
	data, err := os.ReadFile(testJPEG)
	if err != nil {
		t.Fatal(err)
	}
	e, err := exif.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	t.Log(e)
}

func Test_exif_003(t *testing.T) {
	f, err := os.Open(testJPEG)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	e, err := exif.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	t.Log(e)
}

////////////////////////////////////////////////////////////////////////////////
// TAGS

func Test_exif_010(t *testing.T) {
	e, err := exif.Open(testJPEG)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	tags := e.Tags()
	if len(tags) == 0 {
		t.Fatal("expected at least one tag")
	}
	for _, tag := range tags {
		t.Logf("IFD=%d Tag=0x%04X Name=%q Format=%v Components=%d Value=%q Decoded=%v",
			tag.IFD(), tag.Tag(), tag.Name(), tag.Format(), tag.Components(), tag.Value(), tag.Any())
	}
}

func Test_exif_011(t *testing.T) {
	e, err := exif.Open(testJPEG)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	for _, tag := range e.Tags() {
		if tag.Name() == "" {
			t.Errorf("tag 0x%04X has empty name", tag.Tag())
		}
		if tag.String() == "" {
			t.Errorf("tag 0x%04X has empty string value", tag.Tag())
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// JSON

func Test_exif_015(t *testing.T) {
	e, err := exif.Open(testJPEG)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	for _, tag := range e.Tags() {
		data, err := json.Marshal(tag)
		if err != nil {
			t.Fatalf("tag 0x%04X: %v", tag.Tag(), err)
		}
		t.Log(string(data))
	}
}

////////////////////////////////////////////////////////////////////////////////
// MAKERNOTE

func Test_exif_020(t *testing.T) {
	e, err := exif.Open(testJPEG)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if mn := e.MakerNote(); mn != nil {
		t.Fatal("expected no makernote in sample.jpg")
	}
}

func Test_exif_021(t *testing.T) {
	e, err := exif.Open(testJPEGMakerNote)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	mn := e.MakerNote()
	if mn == nil {
		t.Fatal("expected makernote")
	}
	data, err := json.Marshal(mn)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func Test_exif_022(t *testing.T) {
	e, err := exif.Open(testJPEGMakerNote)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func Test_exif_030_heif_payload_shapes(t *testing.T) {
	ctx := libheif.Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer libheif.Libheif_context_free(ctx)

	if err := libheif.Libheif_context_read_from_file(ctx, "../../etc/test/photo.HEIC"); err != nil {
		t.Fatal(err)
	}

	handle, err := libheif.Libheif_context_get_primary_image_handle(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer libheif.Libheif_image_handle_release(handle)

	count := libheif.Libheif_image_handle_get_number_of_metadata_blocks(handle, "Exif")
	if count <= 0 {
		t.Fatal("expected EXIF metadata")
	}

	ids := libheif.Libheif_image_handle_get_list_of_metadata_block_IDs(handle, "Exif", count)
	if len(ids) == 0 {
		t.Fatal("expected EXIF metadata IDs")
	}

	data, err := libheif.Libheif_image_handle_get_metadata(handle, ids[0])
	if err != nil {
		t.Fatal(err)
	}

	for name, payload := range map[string][]byte{
		"full":   data,
		"skip4":  data[4:],
		"skip10": data[10:],
	} {
		doc, err := exif.Parse(payload)
		if err != nil {
			t.Logf("%s: err=%v", name, err)
			continue
		}
		t.Logf("%s: tags=%d", name, len(doc.Tags()))
		_ = doc.Close()
	}
}
