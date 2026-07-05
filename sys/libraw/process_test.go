package libraw_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libraw"
)

func Test_process_000(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}
	if rc := Libraw_unpack_thumb(data); rc != 0 {
		t.Fatalf("Libraw_unpack_thumb: %v", Libraw_strerror(rc))
	}

	thumb := Libraw_get_thumbnail(data)
	if thumb == nil {
		t.Fatal("Libraw_get_thumbnail returned nil")
	}
	t.Logf("thumbnail format=%v size=%dx%d length=%d",
		Thumbnail_format(thumb), Thumbnail_width(thumb), Thumbnail_height(thumb), Thumbnail_length(thumb))

	if Thumbnail_length(thumb) == 0 {
		t.Error("expected non-zero thumbnail length")
	}
	if len(Thumbnail_data(thumb)) == 0 {
		t.Error("expected non-empty thumbnail data")
	}
}

func Test_process_001(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}
	if rc := Libraw_unpack(data); rc != 0 {
		t.Fatalf("Libraw_unpack: %v", Libraw_strerror(rc))
	}
	if rc := Libraw_dcraw_process(data); rc != 0 {
		t.Fatalf("Libraw_dcraw_process: %v", Libraw_strerror(rc))
	}
}

func Test_process_002(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}
	if rc := Libraw_unpack(data); rc != 0 {
		t.Fatalf("Libraw_unpack: %v", Libraw_strerror(rc))
	}

	Libraw_subtract_black(data)

	if rc := Libraw_raw2image(data); rc != 0 {
		t.Fatalf("Libraw_raw2image: %v", Libraw_strerror(rc))
	}
	Libraw_free_image(data)
}

func Test_process_003(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}
	if rc := Libraw_adjust_sizes_info_only(data); rc != 0 {
		t.Fatalf("Libraw_adjust_sizes_info_only: %v", Libraw_strerror(rc))
	}

	w := Libraw_get_iwidth(data)
	h := Libraw_get_iheight(data)
	t.Logf("adjusted image size=%dx%d", w, h)
}

func Test_process_004(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}
	if rc := Libraw_unpack(data); rc != 0 {
		t.Fatalf("Libraw_unpack: %v", Libraw_strerror(rc))
	}
	if rc := Libraw_dcraw_process(data); rc != 0 {
		t.Fatalf("Libraw_dcraw_process: %v", Libraw_strerror(rc))
	}

	t.Logf("cam_mul=[%.4f %.4f %.4f %.4f]",
		Libraw_get_cam_mul(data, 0), Libraw_get_cam_mul(data, 1),
		Libraw_get_cam_mul(data, 2), Libraw_get_cam_mul(data, 3))
	t.Logf("pre_mul=[%.4f %.4f %.4f %.4f]",
		Libraw_get_pre_mul(data, 0), Libraw_get_pre_mul(data, 1),
		Libraw_get_pre_mul(data, 2), Libraw_get_pre_mul(data, 3))
	t.Logf("color_maximum=%d", Libraw_get_color_maximum(data))
}
