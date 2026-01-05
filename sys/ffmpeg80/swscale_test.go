package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST SWSFLAG STRING METHODS

func Test_swscale_flag_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		flag     SWSFlag
		expected string
	}{
		{SWS_NONE, "SWS_NONE"},
		{SWS_FAST_BILINEAR, "SWS_FAST_BILINEAR"},
		{SWS_BILINEAR, "SWS_BILINEAR"},
		{SWS_BICUBIC, "SWS_BICUBIC"},
		{SWS_X, "SWS_X"},
		{SWS_POINT, "SWS_POINT"},
		{SWS_AREA, "SWS_AREA"},
		{SWS_BICUBLIN, "SWS_BICUBLIN"},
		{SWS_GAUSS, "SWS_GAUSS"},
		{SWS_SINC, "SWS_SINC"},
		{SWS_LANCZOS, "SWS_LANCZOS"},
		{SWS_SPLINE, "SWS_SPLINE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.flag.FlagString()
			assert.Equal(tt.expected, result)
			t.Logf("%s => %s", tt.expected, result)
		})
	}
}

func Test_swscale_flag_none_string(t *testing.T) {
	assert := assert.New(t)

	flag := SWS_NONE
	result := flag.String()
	assert.Equal("SWS_NONE", result)
}

func Test_swscale_flag_invalid(t *testing.T) {
	assert := assert.New(t)

	flag := SWSFlag(99999)
	result := flag.FlagString()
	assert.Contains(result, "Invalid")
}

func Test_swscale_context_string(t *testing.T) {
	assert := assert.New(t)

	// Test nil context
	var ctx *SWSContext
	assert.Equal("<nil>", ctx.String(), "nil context should return \"<nil>\"")

	// Test valid context
	ctx = SWScale_get_context(
		320, 240, AV_PIX_FMT_YUV420P,
		640, 480, AV_PIX_FMT_RGB24,
		SWS_BILINEAR, nil, nil, nil,
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	str := ctx.String()
	assert.Equal("<SWSContext>", str, "Context String() should return \"<SWSContext>\"")

	t.Logf("Context string representation: %s", str)
}

func Test_swscale_filter_string(t *testing.T) {
	assert := assert.New(t)

	// Test nil filter
	var filter *SWSFilter
	assert.Equal("<nil>", filter.String(), "nil filter should return \"<nil>\"")

	t.Log("Filter string works for nil")
}
