package libheif_test

import (
	"bytes"
	"testing"

	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_context_write_000(t *testing.T) {
	if !Libheif_have_encoder_for_format(HEIF_COMPRESSION_HEVC) {
		t.Skip("no HEVC encoder available in this libheif build")
	}

	if err := Libheif_init(); err != nil {
		t.Fatalf("Libheif_init error=%v", err)
	}
	defer Libheif_deinit()

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

	encoder, err := Libheif_context_get_encoder_for_format(ctx, HEIF_COMPRESSION_HEVC)
	if err != nil {
		t.Fatalf("Libheif_context_get_encoder_for_format error=%v", err)
	}
	if encoder == nil {
		t.Fatal("Libheif_context_get_encoder_for_format returned nil")
	}
	defer Libheif_encoder_release(encoder)

	if err := Libheif_encoder_set_lossy_quality(encoder, 75); err != nil {
		t.Fatalf("Libheif_encoder_set_lossy_quality error=%v", err)
	}

	handle, err := Libheif_context_encode_image(ctx, img, encoder, nil)
	if err != nil {
		t.Fatalf("Libheif_context_encode_image error=%v", err)
	}
	if handle == nil {
		t.Fatal("Libheif_context_encode_image returned nil handle")
	}
	defer Libheif_image_handle_release(handle)

	if err := Libheif_context_set_primary_image(ctx, handle); err != nil {
		t.Fatalf("Libheif_context_set_primary_image error=%v", err)
	}
	Libheif_context_set_major_brand(ctx, HEIF_BRAND2_HEIC)
	Libheif_context_add_compatible_brand(ctx, HEIF_BRAND2_MIF1)
	Libheif_context_set_write_mini_format(ctx, false)

	var encoded bytes.Buffer
	if err := Libheif_context_write(ctx, func(data []byte) error {
		_, writeErr := encoded.Write(data)
		return writeErr
	}); err != nil {
		t.Fatalf("Libheif_context_write error=%v", err)
	}
	if encoded.Len() == 0 {
		t.Fatal("Libheif_context_write produced no data")
	}

	readCtx := Libheif_context_alloc()
	if readCtx == nil {
		t.Fatal("Libheif_context_alloc returned nil for read context")
	}
	defer Libheif_context_free(readCtx)

	if err := Libheif_context_read_from_memory(readCtx, encoded.Bytes()); err != nil {
		t.Fatalf("Libheif_context_read_from_memory error=%v", err)
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
