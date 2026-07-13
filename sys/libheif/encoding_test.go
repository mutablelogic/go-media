package libheif_test

import (
	"path/filepath"
	"testing"

	// Packages
	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_encoding_000(t *testing.T) {
	if !Libheif_have_encoder_for_format(HEIF_COMPRESSION_HEVC) {
		t.Skip("no HEVC encoder available in this libheif build")
	}

	img, err := Libheif_image_create(2, 2, HEIF_COLORSPACE_RGB, HEIF_CHROMA_INTERLEAVED_RGB)
	if err != nil {
		t.Fatalf("Libheif_image_create error=%v", err)
	}
	defer Libheif_image_release(img)

	if err := Libheif_image_add_plane(img, HEIF_CHANNEL_INTERLEAVED, 2, 2, 8); err != nil {
		t.Fatalf("Libheif_image_add_plane error=%v", err)
	}

	plane, stride := Libheif_image_get_plane(img, HEIF_CHANNEL_INTERLEAVED)
	if len(plane) == 0 || stride < 6 {
		t.Fatalf("invalid writable plane len=%d stride=%d", len(plane), stride)
	}

	copy(plane[0:6], []byte{255, 0, 0, 0, 255, 0})
	copy(plane[stride:stride+6], []byte{0, 0, 255, 255, 255, 255})

	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	enc, err := Libheif_context_get_encoder_for_format(ctx, HEIF_COMPRESSION_HEVC)
	if err != nil {
		t.Fatalf("Libheif_context_get_encoder_for_format error=%v", err)
	}
	if enc == nil {
		t.Fatal("Libheif_context_get_encoder_for_format returned nil")
	}
	defer Libheif_encoder_release(enc)

	if err := Libheif_encoder_set_lossy_quality(enc, 75); err != nil {
		t.Fatalf("Libheif_encoder_set_lossy_quality error=%v", err)
	}

	handle, err := Libheif_context_encode_image(ctx, img, enc, nil)
	if err != nil {
		t.Fatalf("Libheif_context_encode_image error=%v", err)
	}
	if handle == nil {
		t.Fatal("Libheif_context_encode_image returned nil handle")
	}
	defer Libheif_image_handle_release(handle)

	outfile := filepath.Join(t.TempDir(), "roundtrip.heic")
	if err := Libheif_context_write_to_file(ctx, outfile); err != nil {
		t.Fatalf("Libheif_context_write_to_file error=%v", err)
	}

	readCtx := Libheif_context_alloc()
	if readCtx == nil {
		t.Fatal("Libheif_context_alloc returned nil (read context)")
	}
	defer Libheif_context_free(readCtx)

	if err := Libheif_context_read_from_file(readCtx, outfile); err != nil {
		t.Fatalf("Libheif_context_read_from_file error=%v", err)
	}

	decoded, err := Libheif_decode_primary_image_rgb(readCtx)
	if err != nil {
		t.Fatalf("Libheif_decode_primary_image_rgb error=%v", err)
	}
	if decoded == nil {
		t.Fatal("Libheif_decode_primary_image_rgb returned nil")
	}
	defer Libheif_image_release(decoded)

	if gotW := Libheif_image_get_primary_width(decoded); gotW != 2 {
		t.Fatalf("decoded width=%d want=2", gotW)
	}
	if gotH := Libheif_image_get_primary_height(decoded); gotH != 2 {
		t.Fatalf("decoded height=%d want=2", gotH)
	}
}
