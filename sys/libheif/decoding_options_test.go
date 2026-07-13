package libheif_test

import (
	"testing"

	// Packages
	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_decoding_options_000(t *testing.T) {
	opts := Libheif_decoding_options_alloc()
	if opts == nil {
		t.Fatal("Libheif_decoding_options_alloc returned nil")
	}
	defer Libheif_decoding_options_free(opts)

	dup := Libheif_decoding_options_alloc()
	if dup == nil {
		t.Fatal("Libheif_decoding_options_alloc returned nil for duplicate")
	}
	defer Libheif_decoding_options_free(dup)

	Libheif_decoding_options_copy(dup, opts)
}

func Test_decoding_options_001(t *testing.T) {
	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	Libheif_context_set_max_decoding_threads(ctx, 1)
	if got := Libheif_context_get_max_decoding_threads(ctx); got != 1 {
		t.Fatalf("Libheif_context_get_max_decoding_threads=%d want=1", got)
	}
}

func Test_decoding_options_002(t *testing.T) {
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

	opts := Libheif_decoding_options_alloc()
	if opts == nil {
		t.Fatal("Libheif_decoding_options_alloc returned nil")
	}
	defer Libheif_decoding_options_free(opts)

	img, err := Libheif_decode_image_with_options(handle, HEIF_COLORSPACE_RGB, HEIF_CHROMA_INTERLEAVED_RGB, opts)
	if err != nil {
		t.Fatalf("Libheif_decode_image_with_options error=%v", err)
	}
	if img == nil {
		t.Fatal("Libheif_decode_image_with_options returned nil image")
	}
	defer Libheif_image_release(img)
}
