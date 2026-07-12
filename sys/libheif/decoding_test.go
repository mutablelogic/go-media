package libheif_test

import (
	"testing"

	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_decoding_000(t *testing.T) {
	requireHEVCDecoder(t)

	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	if err := Libheif_context_read_from_file(ctx, testHEIC); err != nil {
		t.Fatalf("Libheif_context_read_from_file error=%v", err)
	}

	handle, err := Libheif_context_get_primary_image_handle(ctx)
	if err != nil {
		t.Fatalf("Libheif_context_get_primary_image_handle error=%v", err)
	}
	if handle == nil {
		t.Fatal("Libheif_context_get_primary_image_handle returned nil handle")
	}
	defer Libheif_image_handle_release(handle)

	img, err := Libheif_decode_image(handle, HEIF_COLORSPACE_RGB, HEIF_CHROMA_INTERLEAVED_RGB)
	if err != nil {
		t.Fatalf("Libheif_decode_image error=%v", err)
	}
	if img == nil {
		t.Fatal("Libheif_decode_image returned nil image")
	}
	defer Libheif_image_release(img)

	if w := Libheif_image_get_primary_width(img); w <= 0 {
		t.Fatalf("Libheif_image_get_primary_width=%d", w)
	}
	if h := Libheif_image_get_primary_height(img); h <= 0 {
		t.Fatalf("Libheif_image_get_primary_height=%d", h)
	}
	if !Libheif_image_has_channel(img, HEIF_CHANNEL_INTERLEAVED) {
		t.Fatal("decoded image missing interleaved channel")
	}
	plane, stride := Libheif_image_get_plane_readonly(img, HEIF_CHANNEL_INTERLEAVED)
	if len(plane) == 0 || stride <= 0 {
		t.Fatalf("invalid plane data len=%d stride=%d", len(plane), stride)
	}
}

func Test_decoding_001(t *testing.T) {
	requireHEVCDecoder(t)

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
		t.Fatal("Libheif_decode_primary_image_rgb returned nil image")
	}
	defer Libheif_image_release(img)

	if w := Libheif_image_get_primary_width(img); w <= 0 {
		t.Fatalf("Libheif_image_get_primary_width=%d", w)
	}
}

func Test_decoding_002(t *testing.T) {
	requireHEVCDecoder(t)

	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	if err := Libheif_context_read_from_file(ctx, testHEIC); err != nil {
		t.Fatalf("Libheif_context_read_from_file error=%v", err)
	}

	id, err := Libheif_context_get_primary_image_ID(ctx)
	if err != nil {
		t.Fatalf("Libheif_context_get_primary_image_ID error=%v", err)
	}

	img, err := Libheif_decode_image_by_item_id(ctx, id, HEIF_COLORSPACE_RGB, HEIF_CHROMA_INTERLEAVED_RGB)
	if err != nil {
		t.Fatalf("Libheif_decode_image_by_item_id error=%v", err)
	}
	if img == nil {
		t.Fatal("Libheif_decode_image_by_item_id returned nil image")
	}
	defer Libheif_image_release(img)

	if h := Libheif_image_get_primary_height(img); h <= 0 {
		t.Fatalf("Libheif_image_get_primary_height=%d", h)
	}
}
