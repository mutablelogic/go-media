package media_test

import (
	"testing"

	// Namespace import
	. "github.com/djthorpe/go-media/pkg/media"
)

func Test_Util_001(t *testing.T) {
	tests := []struct {
		in   string
		n, m uint
	}{
		{"1", 1, 0},
		{"1/", 0, 0},
		{"1/1", 1, 1},
		{"/1", 0, 0},
		{"101", 101, 0},
		{"101/102", 101, 102},
	}
	for _, test := range tests {
		if n, m := ParseTrackDisc(test.in); n != test.n || m != test.m {
			t.Errorf("Unexpected return for %q: n=%v m=%v", test.in, n, m)
		}
	}
}
