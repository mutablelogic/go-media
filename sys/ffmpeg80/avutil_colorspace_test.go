package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_avutil_colorspace_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		space    AVColorSpace
		expected string
	}{
		{AVCOL_SPC_RGB, "AVCOL_SPC_RGB"},
		{AVCOL_SPC_BT709, "AVCOL_SPC_BT709"},
		{AVCOL_SPC_UNSPECIFIED, "AVCOL_SPC_UNSPECIFIED"},
		{AVCOL_SPC_RESERVED, "AVCOL_SPC_RESERVED"},
		{AVCOL_SPC_FCC, "AVCOL_SPC_FCC"},
		{AVCOL_SPC_BT470BG, "AVCOL_SPC_BT470BG"},
		{AVCOL_SPC_SMPTE170M, "AVCOL_SPC_SMPTE170M"},
		{AVCOL_SPC_SMPTE240M, "AVCOL_SPC_SMPTE240M"},
		{AVCOL_SPC_YCGCO, "AVCOL_SPC_YCGCO"},
		{AVCOL_SPC_BT2020_NCL, "AVCOL_SPC_BT2020_NCL"},
		{AVCOL_SPC_BT2020_CL, "AVCOL_SPC_BT2020_CL"},
		{AVCOL_SPC_SMPTE2085, "AVCOL_SPC_SMPTE2085"},
		{AVCOL_SPC_CHROMA_DERIVED_NCL, "AVCOL_SPC_CHROMA_DERIVED_NCL"},
		{AVCOL_SPC_CHROMA_DERIVED_CL, "AVCOL_SPC_CHROMA_DERIVED_CL"},
		{AVCOL_SPC_ICTCP, "AVCOL_SPC_ICTCP"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.space.String()
			assert.Equal(tt.expected, result)
			t.Logf("%s => %s", tt.expected, result)
		})
	}
}

func Test_avutil_colorspace_invalid(t *testing.T) {
	assert := assert.New(t)

	space := AVColorSpace(999)
	result := space.String()
	assert.Contains(result, "AVColorSpace")
	assert.Contains(result, "999")

	t.Logf("Invalid colorspace string: %s", result)
}

func Test_avutil_colorrange_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		range_   AVColorRange
		expected string
	}{
		{AVCOL_RANGE_UNSPECIFIED, "AVCOL_RANGE_UNSPECIFIED"},
		{AVCOL_RANGE_MPEG, "AVCOL_RANGE_MPEG"},
		{AVCOL_RANGE_JPEG, "AVCOL_RANGE_JPEG"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.range_.String()
			assert.Equal(tt.expected, result)
			t.Logf("%s => %s", tt.expected, result)
		})
	}
}

func Test_avutil_colorrange_invalid(t *testing.T) {
	assert := assert.New(t)

	range_ := AVColorRange(999)
	result := range_.String()
	assert.Contains(result, "AVColorRange")
	assert.Contains(result, "999")

	t.Logf("Invalid color range string: %s", result)
}

func Test_avutil_frame_colorspace_getters(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	if !assert.NotNil(frame) {
		t.FailNow()
	}
	defer AVUtil_frame_free(frame)

	// Set colorspace
	frame.SetColorspace(AVCOL_SPC_BT709)
	assert.Equal(AVCOL_SPC_BT709, frame.Colorspace())
	t.Logf("Frame colorspace: %s", frame.Colorspace())

	// Set color range
	frame.SetColorRange(AVCOL_RANGE_MPEG)
	assert.Equal(AVCOL_RANGE_MPEG, frame.ColorRange())
	t.Logf("Frame color range: %s", frame.ColorRange())
}
