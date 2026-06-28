package libexif_test

import (
	"os"
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

func Test_data_000(t *testing.T) {
	data := Exif_data_new()
	if data == nil {
		t.Fatal("Exif_data_new returned nil")
	}
	defer Exif_data_unref(data)
	t.Log("data=", data)
}

func Test_data_001(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)
	t.Log("data=", data)
}

func Test_data_002(t *testing.T) {
	raw, err := os.ReadFile(testJPEG)
	if err != nil {
		t.Fatal(err)
	}

	data := Exif_data_new_from_data(raw)
	if data == nil {
		t.Fatal("Exif_data_new_from_data returned nil")
	}
	defer Exif_data_unref(data)
	t.Log("data=", data)
}

func Test_data_003(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	order := Exif_data_get_byte_order(data)
	name := Exif_byte_order_get_name(order)
	if name == "" {
		t.Fatal("expected non-empty byte order name")
	}
	t.Log("byte order=", name)
}

func Test_data_004(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	Exif_data_set_byte_order(data, EXIF_BYTE_ORDER_INTEL)
	if got := Exif_data_get_byte_order(data); got != EXIF_BYTE_ORDER_INTEL {
		t.Fatalf("expected INTEL byte order, got %v", got)
	}

	Exif_data_set_byte_order(data, EXIF_BYTE_ORDER_MOTOROLA)
	if got := Exif_data_get_byte_order(data); got != EXIF_BYTE_ORDER_MOTOROLA {
		t.Fatalf("expected MOTOROLA byte order, got %v", got)
	}
}

func Test_data_005(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	dtype := Exif_data_get_data_type(data)
	t.Log("data type=", dtype)
}

func Test_data_006(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	Exif_data_set_data_type(data, EXIF_DATA_TYPE_COMPRESSED)
	if got := Exif_data_get_data_type(data); got != EXIF_DATA_TYPE_COMPRESSED {
		t.Fatalf("expected COMPRESSED data type, got %v", got)
	}
}

func Test_data_007(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	raw := Exif_data_save_data(data)
	if len(raw) == 0 {
		t.Fatal("Exif_data_save_data returned empty bytes")
	}
	t.Log("saved bytes=", len(raw))

	data2 := Exif_data_new()
	if data2 == nil {
		t.Fatal("Exif_data_new returned nil")
	}
	defer Exif_data_unref(data2)

	Exif_data_load_data(data2, raw)
	raw2 := Exif_data_save_data(data2)
	if len(raw2) == 0 {
		t.Fatal("expected non-empty bytes after load_data + save_data")
	}
	t.Log("round-tripped bytes=", len(raw2))
}

func Test_data_008(t *testing.T) {
	data := Exif_data_new()
	if data == nil {
		t.Fatal("Exif_data_new returned nil")
	}
	defer Exif_data_unref(data)

	for _, opt := range []DataOption{
		EXIF_DATA_OPTION_IGNORE_UNKNOWN_TAGS,
		EXIF_DATA_OPTION_FOLLOW_SPECIFICATION,
		EXIF_DATA_OPTION_DONT_CHANGE_MAKER_NOTE,
	} {
		name := Exif_data_option_get_name(data, opt)
		desc := Exif_data_option_get_description(data, opt)
		if name == "" {
			t.Errorf("option %v: expected non-empty name", opt)
		}
		if desc == "" {
			t.Errorf("option %v: expected non-empty description", opt)
		}
		t.Logf("option %v: name=%q desc=%q", opt, name, desc)
	}
}

func Test_data_009(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	Exif_data_set_option(data, EXIF_DATA_OPTION_IGNORE_UNKNOWN_TAGS)
	Exif_data_unset_option(data, EXIF_DATA_OPTION_IGNORE_UNKNOWN_TAGS)
}

func Test_data_010(t *testing.T) {
	data := Exif_data_new_from_file(testJPEGMakerNote)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	mnote := Exif_data_get_maker_note_data(data)
	if mnote == nil {
		t.Fatal("expected non-nil maker note data")
	}
	t.Log("maker note data=", mnote)
}

func Test_data_011(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	Exif_data_fix(data)
}
