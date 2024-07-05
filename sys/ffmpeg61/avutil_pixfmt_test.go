package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avutil_pixfmt_002(t *testing.T) {
	var opaque uintptr
	for {
		fmt := AVUtil_next_pixel_fmt(&opaque)
		if fmt == AV_PIX_FMT_NONE {
			break
		}
		t.Logf("pixel_fmt[%d]=%v", fmt, AVUtil_get_pix_fmt_name(fmt))
	}
}
