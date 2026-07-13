package libheif_test

import (
	"testing"

	// Packages
	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_image_handle_000(t *testing.T) {
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

	if !Libheif_image_handle_is_primary_image(handle) {
		t.Fatal("primary image handle is not marked as primary")
	}

	if w := Libheif_image_handle_get_width(handle); w <= 0 {
		t.Fatalf("Libheif_image_handle_get_width=%d", w)
	}
	if h := Libheif_image_handle_get_height(handle); h <= 0 {
		t.Fatalf("Libheif_image_handle_get_height=%d", h)
	}

	id := Libheif_image_handle_get_item_id(handle)
	if id == 0 {
		t.Fatal("Libheif_image_handle_get_item_id returned 0")
	}
}

func Test_image_handle_001(t *testing.T) {
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

	n := Libheif_image_handle_get_number_of_metadata_blocks(handle, "")
	if n < 0 {
		t.Fatalf("Libheif_image_handle_get_number_of_metadata_blocks=%d", n)
	}
	if n == 0 {
		return
	}

	ids := Libheif_image_handle_get_list_of_metadata_block_IDs(handle, "", n)
	if len(ids) == 0 {
		t.Fatal("Libheif_image_handle_get_list_of_metadata_block_IDs returned no IDs")
	}

	first := ids[0]
	if sz := Libheif_image_handle_get_metadata_size(handle, first); sz < 0 {
		t.Fatalf("Libheif_image_handle_get_metadata_size=%d", sz)
	}

	_, _ = Libheif_image_handle_get_metadata_type(handle, first), Libheif_image_handle_get_metadata_content_type(handle, first)
}
