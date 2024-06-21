package ffmpeg_test

import (
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avutil_samplefmt_000(t *testing.T) {
	assert := assert.New(t)
	for fmt := AV_SAMPLE_FMT_NONE; fmt < AV_SAMPLE_FMT_NB; fmt++ {
		if fmt == AV_SAMPLE_FMT_NONE {
			continue
		}
		t.Logf("sample_fmt[%d]=%v", fmt, AVUtil_get_sample_fmt_name(fmt))
		fmt_ := AVUtil_get_sample_fmt(AVUtil_get_sample_fmt_name(fmt))
		assert.Equal(fmt, fmt_)
	}
}

func Test_avutil_samplefmt_001(t *testing.T) {
	for fmt := AV_SAMPLE_FMT_NONE; fmt < AV_SAMPLE_FMT_NB; fmt++ {
		if fmt == AV_SAMPLE_FMT_NONE {
			continue
		}
		t.Logf("sample_fmt[%d]=%v", fmt, AVUtil_get_sample_fmt_name(fmt))
		t.Logf("  is_planar=%v", AVUtil_sample_fmt_is_planar(fmt))
		t.Logf("  bytes_per_sample=%v", AVUtil_get_bytes_per_sample(fmt))
	}
}

func Test_avutil_samplefmt_002(t *testing.T) {
	var opaque uintptr
	for {
		fmt := AVUtil_next_sample_fmt(&opaque)
		if fmt == AV_SAMPLE_FMT_NONE {
			break
		}
		t.Logf("sample_fmt[%d]=%v", fmt, AVUtil_get_sample_fmt_name(fmt))
	}
}
