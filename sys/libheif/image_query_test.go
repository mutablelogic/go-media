package libheif_test

import (
	"testing"

	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_image_query_000(t *testing.T) {
	img, err := Libheif_image_create(3, 2, HEIF_COLORSPACE_RGB, HEIF_CHROMA_INTERLEAVED_RGB)
	if err != nil {
		t.Fatalf("Libheif_image_create error=%v", err)
	}
	defer Libheif_image_release(img)

	Libheif_image_set_pixel_aspect_ratio(img, 2, 1)
	aspectH, aspectV := Libheif_image_get_pixel_aspect_ratio(img)
	if aspectH != 2 || aspectV != 1 {
		t.Fatalf("pixel aspect ratio=%d:%d want=2:1", aspectH, aspectV)
	}
}

func Test_image_query_001(t *testing.T) {
	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	if err := Libheif_context_read_from_file(ctx, testHEIC); err != nil {
		t.Fatalf("Libheif_context_read_from_file error=%v", err)
	}

	img, err := Libheif_decode_primary_image_rgb(ctx)
	if err != nil {
		t.Fatalf("Libheif_decode_primary_image_rgb error=%v", err)
	}
	if img == nil {
		t.Fatal("Libheif_decode_primary_image_rgb returned nil")
	}
	defer Libheif_image_release(img)

	warnings, count := Libheif_image_get_decoding_warnings(img, 0, 0)
	if count < 0 {
		t.Fatalf("Libheif_image_get_decoding_warnings count=%d", count)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for fixture decode, got %d", len(warnings))
	}
}
