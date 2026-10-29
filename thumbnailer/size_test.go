package thumbnailer

import (
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
)

func TestSize(t *testing.T) {
	// An anamorphic video stored at 540x576, displayed at 768x576, as its
	// first frame's sample aspect ratio gives
	th := &Thumbnailer{width: 768, height: 576, displayed: true}

	// Thumbnails keep the displayed aspect ratio, with even sizes for 4:2:0
	for _, test := range []struct {
		width, height int
		want          [2]int
	}{
		{0, 0, [2]int{768, 576}},     // the displayed size
		{640, 0, [2]int{640, 480}},   // 4:3
		{320, 0, [2]int{320, 240}},   // as frames are measured
		{0, 240, [2]int{320, 240}},   // from the height
		{1920, 0, [2]int{768, 576}},  // no larger than displayed
		{200, 100, [2]int{200, 100}}, // both given, as they are
		{101, 0, [2]int{100, 76}},    // rounded, then down to even
	} {
		w, h := th.size(test.width, test.height)
		assert.Equalf(t, test.want, [2]int{w, h}, "%dx%d", test.width, test.height)
	}
}
