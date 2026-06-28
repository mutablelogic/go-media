package libexif_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

var allFormats = []Format{
	EXIF_FORMAT_BYTE,
	EXIF_FORMAT_ASCII,
	EXIF_FORMAT_SHORT,
	EXIF_FORMAT_LONG,
	EXIF_FORMAT_RATIONAL,
	EXIF_FORMAT_SBYTE,
	EXIF_FORMAT_UNDEFINED,
	EXIF_FORMAT_SSHORT,
	EXIF_FORMAT_SLONG,
	EXIF_FORMAT_SRATIONAL,
	EXIF_FORMAT_FLOAT,
	EXIF_FORMAT_DOUBLE,
}

func Test_format_000(t *testing.T) {
	for _, f := range allFormats {
		name := Exif_format_get_name(f)
		if name == "" {
			t.Errorf("format %d: empty name", f)
		}
		t.Logf("format %d: name=%q", f, name)
	}
}

func Test_format_001(t *testing.T) {
	for _, f := range allFormats {
		size := Exif_format_get_size(f)
		if size == 0 {
			t.Errorf("format %d (%s): expected non-zero size", f, Exif_format_get_name(f))
		}
		t.Logf("format %d (%s): size=%d bytes", f, Exif_format_get_name(f), size)
	}
}

func Test_format_002(t *testing.T) {
	// spot-check known sizes
	cases := []struct {
		format Format
		size   uint
	}{
		{EXIF_FORMAT_BYTE, 1},
		{EXIF_FORMAT_ASCII, 1},
		{EXIF_FORMAT_SHORT, 2},
		{EXIF_FORMAT_LONG, 4},
		{EXIF_FORMAT_RATIONAL, 8},
		{EXIF_FORMAT_SBYTE, 1},
		{EXIF_FORMAT_SSHORT, 2},
		{EXIF_FORMAT_SLONG, 4},
		{EXIF_FORMAT_SRATIONAL, 8},
		{EXIF_FORMAT_FLOAT, 4},
		{EXIF_FORMAT_DOUBLE, 8},
	}
	for _, tc := range cases {
		got := Exif_format_get_size(tc.format)
		if got != tc.size {
			t.Errorf("format %s: size=%d, want %d", Exif_format_get_name(tc.format), got, tc.size)
		}
	}
}
