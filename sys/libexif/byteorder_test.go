package libexif_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

func Test_byteorder_000(t *testing.T) {
	for _, tc := range []struct {
		order ByteOrder
		want  string
	}{
		{EXIF_BYTE_ORDER_MOTOROLA, "Motorola"},
		{EXIF_BYTE_ORDER_INTEL, "Intel"},
	} {
		got := Exif_byte_order_get_name(tc.order)
		if got != tc.want {
			t.Errorf("byte order %v: got %q, want %q", tc.order, got, tc.want)
		}
		t.Logf("byte order %v: name=%q", tc.order, got)
	}
}

func Test_datatype_000(t *testing.T) {
	for _, dtype := range []DataType{
		EXIF_DATA_TYPE_UNCOMPRESSED_CHUNKY,
		EXIF_DATA_TYPE_UNCOMPRESSED_PLANAR,
		EXIF_DATA_TYPE_UNCOMPRESSED_YCC,
		EXIF_DATA_TYPE_COMPRESSED,
		EXIF_DATA_TYPE_UNKNOWN,
	} {
		t.Logf("data type %d", dtype)
	}
}
