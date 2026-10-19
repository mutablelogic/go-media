package exif

import (
	"testing"

	libexif "github.com/mutablelogic/go-media/sys/libexif"
)

func Test_tag_ascii(t *testing.T) {
	for _, tc := range []struct {
		data, want string
	}{
		{"Canon\x00", "Canon"},
		{"Canon", "Canon"},
		{"\x00", ""},

		// A fixed-size field, padded with null bytes, as written by a
		// Canon PowerShot G1
		{"Canon PowerShot G1\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00", "Canon PowerShot G1"},

		// Anything after the first null byte is padding
		{"Canon\x00junk\x00", "Canon"},
	} {
		tag := &Tag{format: libexif.EXIF_FORMAT_ASCII, components: uint(len(tc.data)), data: []byte(tc.data)}
		if got := tag.Any(); got != tc.want {
			t.Errorf("Any() of %q = %q, want %q", tc.data, got, tc.want)
		}
	}
}
