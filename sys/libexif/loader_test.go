package libexif_test

import (
	"os"
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

const (
	testJPEG         = "../../etc/test/sample.jpg"
	testJPEGMakerNote = "../../etc/test/canon_makernote_variant_1.jpg"
)

var testJPEGs = []string{
	testJPEG,
	"../../etc/test/canon_makernote_variant_1.jpg",
	"../../etc/test/fuji_makernote_variant_1.jpg",
	"../../etc/test/olympus_makernote_variant_2.jpg",
	"../../etc/test/olympus_makernote_variant_3.jpg",
	"../../etc/test/olympus_makernote_variant_4.jpg",
	"../../etc/test/olympus_makernote_variant_5.jpg",
	"../../etc/test/pentax_makernote_variant_2.jpg",
	"../../etc/test/pentax_makernote_variant_3.jpg",
	"../../etc/test/pentax_makernote_variant_4.jpg",
}

func Test_loader_000(t *testing.T) {
	loader := Exif_loader_new()
	if loader == nil {
		t.Fatal("Exif_loader_new returned nil")
	}
	defer Exif_loader_unref(loader)
	t.Log("loader=", loader)
}

func Test_loader_001(t *testing.T) {
	loader := Exif_loader_new()
	if loader == nil {
		t.Fatal("Exif_loader_new returned nil")
	}
	defer Exif_loader_unref(loader)

	Exif_loader_ref(loader)
	Exif_loader_unref(loader)
}

func Test_loader_002(t *testing.T) {
	loader := Exif_loader_new()
	if loader == nil {
		t.Fatal("Exif_loader_new returned nil")
	}
	defer Exif_loader_unref(loader)

	Exif_loader_write_file(loader, testJPEG)

	buf := Exif_loader_get_buf(loader)
	if len(buf) == 0 {
		t.Fatal("expected non-empty buffer after writing file")
	}
	t.Log("buffer size=", len(buf))
}

func Test_loader_003(t *testing.T) {
	loader := Exif_loader_new()
	if loader == nil {
		t.Fatal("Exif_loader_new returned nil")
	}
	defer Exif_loader_unref(loader)

	Exif_loader_write_file(loader, testJPEG)

	data := Exif_loader_get_data(loader)
	if data == nil {
		t.Fatal("Exif_loader_get_data returned nil")
	}
	defer Exif_data_unref(data)
	t.Log("exif data=", data)
}

func Test_loader_004(t *testing.T) {
	data, err := os.ReadFile(testJPEG)
	if err != nil {
		t.Fatal(err)
	}

	loader := Exif_loader_new()
	if loader == nil {
		t.Fatal("Exif_loader_new returned nil")
	}
	defer Exif_loader_unref(loader)

	Exif_loader_write(loader, data)
	buf := Exif_loader_get_buf(loader)
	if len(buf) == 0 {
		t.Fatal("expected non-empty buffer after write from bytes")
	}
	t.Log("buffer size after write from bytes=", len(buf))
}

func Test_loader_005(t *testing.T) {
	for _, path := range testJPEGs {
		loader := Exif_loader_new()
		if loader == nil {
			t.Fatalf("%s: Exif_loader_new returned nil", path)
		}
		Exif_loader_write_file(loader, path)
		data := Exif_loader_get_data(loader)
		Exif_loader_unref(loader)
		if data == nil {
			t.Errorf("%s: Exif_loader_get_data returned nil", path)
			continue
		}
		Exif_data_unref(data)
		t.Logf("%s: ok", path)
	}
}
