package libraw_test

import (
	"os"
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libraw"
)

func Test_open_000(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)
}

func Test_open_001(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}
}

func Test_open_002(t *testing.T) {
	buf, err := os.ReadFile(testRAW)
	if err != nil {
		t.Fatal(err)
	}

	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_buffer(data, buf); rc != 0 {
		t.Fatalf("Libraw_open_buffer: %v", Libraw_strerror(rc))
	}
}

func Test_open_003(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}

	p := Libraw_get_iparams(data)
	if p == nil {
		t.Fatal("Libraw_get_iparams returned nil")
	}
	t.Logf("make=%q model=%q raw_count=%d colors=%d",
		IParams_make(p), IParams_model(p), IParams_raw_count(p), IParams_colors(p))

	if IParams_make(p) == "" {
		t.Error("expected non-empty camera make")
	}
	if IParams_model(p) == "" {
		t.Error("expected non-empty camera model")
	}
}

func Test_open_004(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}

	other := Libraw_get_imgother(data)
	if other == nil {
		t.Fatal("Libraw_get_imgother returned nil")
	}
	t.Logf("iso_speed=%.0f shutter=%.6f aperture=%.1f focal_len=%.1f timestamp=%d",
		ImgOther_iso_speed(other), ImgOther_shutter(other),
		ImgOther_aperture(other), ImgOther_focal_len(other),
		ImgOther_timestamp(other))
}

func Test_open_005(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}

	w := Libraw_get_raw_width(data)
	h := Libraw_get_raw_height(data)
	iw := Libraw_get_iwidth(data)
	ih := Libraw_get_iheight(data)
	t.Logf("raw=%dx%d image=%dx%d", w, h, iw, ih)

	if w == 0 || h == 0 {
		t.Error("expected non-zero raw dimensions")
	}
}

func Test_open_006(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}

	sizes := Libraw_get_sizes(data)
	if sizes == nil {
		t.Fatal("Libraw_get_sizes returned nil")
	}
}

func Test_open_007(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}

	lens := Libraw_get_lensinfo(data)
	if lens == nil {
		t.Fatal("Libraw_get_lensinfo returned nil")
	}
}

func Test_open_008(t *testing.T) {
	data := Libraw_init(0)
	if data == nil {
		t.Fatal("Libraw_init returned nil")
	}
	defer Libraw_close(data)

	if rc := Libraw_open_file(data, testRAW); rc != 0 {
		t.Fatalf("Libraw_open_file: %v", Libraw_strerror(rc))
	}

	Libraw_recycle_datastream(data)
	Libraw_recycle(data)
}
