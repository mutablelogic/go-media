package libheif_test

import (
	"errors"
	"path/filepath"
	"testing"

	. "github.com/mutablelogic/go-media/sys/libheif"
)

const testHEIC = "../../etc/test/photo.HEIC"

func requireHEVCDecoder(t *testing.T) {
	t.Helper()
	if !Libheif_have_decoder_for_format(HEIF_COMPRESSION_HEVC) {
		t.Skip("no HEVC decoder available in this libheif build")
	}
}

func Test_context_000(t *testing.T) {
	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	Libheif_context_free(ctx)
}

func Test_context_001(t *testing.T) {
	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	path := filepath.Join(t.TempDir(), "missing.heic")
	err := Libheif_context_read_from_file(ctx, path)
	if err == nil {
		t.Fatal("Libheif_context_read_from_file returned nil error for missing file")
	}
	var heifErr HeifError
	if !errors.As(err, &heifErr) {
		t.Fatalf("expected HeifError, got %T (%v)", err, err)
	}
	if heifErr.Code != HEIF_ERROR_INPUT_DOES_NOT_EXIST {
		t.Fatalf("Libheif_context_read_from_file code=%d want=%d message=%q", heifErr.Code, HEIF_ERROR_INPUT_DOES_NOT_EXIST, heifErr.Message)
	}
}

func Test_context_002(t *testing.T) {
	ctx := Libheif_context_alloc()
	if ctx == nil {
		t.Fatal("Libheif_context_alloc returned nil")
	}
	defer Libheif_context_free(ctx)

	err := Libheif_context_read_from_file(ctx, testHEIC)
	if err != nil {
		t.Fatalf("Libheif_context_read_from_file error=%v", err)
	}

	count := Libheif_context_get_number_of_top_level_images(ctx)
	if count <= 0 {
		t.Fatalf("Libheif_context_get_number_of_top_level_images=%d", count)
	}

	ids := Libheif_context_get_list_of_top_level_image_IDs(ctx, count)
	if len(ids) == 0 {
		t.Fatal("Libheif_context_get_list_of_top_level_image_IDs returned no IDs")
	}

	primaryID, err := Libheif_context_get_primary_image_ID(ctx)
	if err != nil {
		t.Fatalf("Libheif_context_get_primary_image_ID error=%v", err)
	}
	if !Libheif_context_is_top_level_image_ID(ctx, primaryID) {
		t.Fatalf("primary image id %d is not reported as top-level", primaryID)
	}
}
