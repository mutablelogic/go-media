package libexif_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

// loadEntry returns the raw data, byte order, and format for a tag in the
// sample JPEG, or skips the test if the tag is not present.
func loadEntry(t *testing.T, ifd IFD, tag Tag) ([]byte, ByteOrder, Format) {
	t.Helper()
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	t.Cleanup(func() { Exif_data_unref(data) })

	order := Exif_data_get_byte_order(data)
	content := Exif_data_get_content(data, ifd)
	if content == nil {
		t.Skipf("no content for IFD %v", ifd)
	}
	entry := Exif_content_get_entry(content, tag)
	if entry == nil {
		t.Skipf("tag 0x%04x not found in IFD %v", tag, ifd)
	}
	raw := Exif_entry_get_data(entry)
	if len(raw) == 0 {
		t.Skipf("tag 0x%04x has no raw data", tag)
	}
	return raw, order, Exif_entry_get_format(entry)
}

// Test_utils_get_short decodes Orientation from IFD0 (should be 1 = top-left).
func Test_utils_get_short(t *testing.T) {
	raw, order, format := loadEntry(t, EXIF_IFD_0, EXIF_TAG_ORIENTATION)
	if format != EXIF_FORMAT_SHORT {
		t.Fatalf("expected SHORT format, got %s", Exif_format_get_name(format))
	}
	got := Exif_get_short(raw, order)
	t.Logf("Orientation raw=%v order=%v value=%d", raw, Exif_byte_order_get_name(order), got)
	if got == 0 {
		t.Fatal("expected non-zero orientation value")
	}
}

// Test_utils_get_rational decodes XResolution from IFD0 (e.g. 72/1).
func Test_utils_get_rational(t *testing.T) {
	raw, order, format := loadEntry(t, EXIF_IFD_0, EXIF_TAG_X_RESOLUTION)
	if format != EXIF_FORMAT_RATIONAL {
		t.Fatalf("expected RATIONAL format, got %s", Exif_format_get_name(format))
	}
	got := Exif_get_rational(raw, order)
	t.Logf("XResolution = %d/%d", got.Numerator, got.Denominator)
	if got.Denominator == 0 {
		t.Fatal("rational denominator is zero")
	}
}

// Test_utils_get_sshort round-trips an int16 value through set/get.
func Test_utils_get_sshort(t *testing.T) {
	buf := make([]byte, 2)
	const want int16 = -42
	Exif_set_sshort(buf, EXIF_BYTE_ORDER_INTEL, want)
	got := Exif_get_sshort(buf, EXIF_BYTE_ORDER_INTEL)
	if got != want {
		t.Fatalf("sshort round-trip: got %d, want %d", got, want)
	}
}

// Test_utils_set_short verifies byte order is respected.
func Test_utils_set_short(t *testing.T) {
	const value uint16 = 0x1234
	buf := make([]byte, 2)

	Exif_set_short(buf, EXIF_BYTE_ORDER_INTEL, value)
	if buf[0] != 0x34 || buf[1] != 0x12 {
		t.Fatalf("Intel: expected [0x34 0x12], got %v", buf)
	}

	Exif_set_short(buf, EXIF_BYTE_ORDER_MOTOROLA, value)
	if buf[0] != 0x12 || buf[1] != 0x34 {
		t.Fatalf("Motorola: expected [0x12 0x34], got %v", buf)
	}

	// round-trip
	got := Exif_get_short(buf, EXIF_BYTE_ORDER_MOTOROLA)
	if got != value {
		t.Fatalf("short round-trip: got %d, want %d", got, value)
	}
}

func Test_utils_set_long(t *testing.T) {
	const value uint32 = 0xDEADBEEF
	buf := make([]byte, 4)

	Exif_set_long(buf, EXIF_BYTE_ORDER_INTEL, value)
	got := Exif_get_long(buf, EXIF_BYTE_ORDER_INTEL)
	if got != value {
		t.Fatalf("long round-trip (Intel): got 0x%x, want 0x%x", got, value)
	}

	Exif_set_long(buf, EXIF_BYTE_ORDER_MOTOROLA, value)
	got = Exif_get_long(buf, EXIF_BYTE_ORDER_MOTOROLA)
	if got != value {
		t.Fatalf("long round-trip (Motorola): got 0x%x, want 0x%x", got, value)
	}
}

func Test_utils_set_slong(t *testing.T) {
	const value int32 = -123456
	buf := make([]byte, 4)
	for _, order := range []ByteOrder{EXIF_BYTE_ORDER_INTEL, EXIF_BYTE_ORDER_MOTOROLA} {
		Exif_set_slong(buf, order, value)
		got := Exif_get_slong(buf, order)
		if got != value {
			t.Errorf("slong round-trip (%v): got %d, want %d", order, got, value)
		}
	}
}

func Test_utils_set_rational(t *testing.T) {
	value := Rational{Numerator: 72, Denominator: 1}
	buf := make([]byte, 8)
	for _, order := range []ByteOrder{EXIF_BYTE_ORDER_INTEL, EXIF_BYTE_ORDER_MOTOROLA} {
		Exif_set_rational(buf, order, value)
		got := Exif_get_rational(buf, order)
		if got != value {
			t.Errorf("rational round-trip (%v): got %+v, want %+v", order, got, value)
		}
	}
}

func Test_utils_set_srational(t *testing.T) {
	value := SRational{Numerator: -1, Denominator: 3}
	buf := make([]byte, 8)
	for _, order := range []ByteOrder{EXIF_BYTE_ORDER_INTEL, EXIF_BYTE_ORDER_MOTOROLA} {
		Exif_set_srational(buf, order, value)
		got := Exif_get_srational(buf, order)
		if got != value {
			t.Errorf("srational round-trip (%v): got %+v, want %+v", order, got, value)
		}
	}
}
