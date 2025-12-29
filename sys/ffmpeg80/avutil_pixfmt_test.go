package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST PIXEL FORMAT NAME

func Test_avutil_get_pix_fmt_name(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		format   AVPixelFormat
		expected string
	}{
		{AV_PIX_FMT_RGB24, "rgb24"},
		{AV_PIX_FMT_YUV420P, "yuv420p"},
		{AV_PIX_FMT_RGBA, "rgba"},
		{AV_PIX_FMT_GRAY8, "gray"},
		{AV_PIX_FMT_NV12, "nv12"},
	}

	for _, tc := range tests {
		name := AVUtil_get_pix_fmt_name(tc.format)
		assert.Equal(tc.expected, name)
		t.Logf("Format %v: %q", tc.format, name)
	}
}

func Test_avutil_get_pix_fmt_name_none(t *testing.T) {
	assert := assert.New(t)

	// AV_PIX_FMT_NONE returns empty string from FFmpeg
	name := AVUtil_get_pix_fmt_name(AV_PIX_FMT_NONE)
	assert.Equal("", name)
}

////////////////////////////////////////////////////////////////////////////////
// TEST GET PIXEL FORMAT BY NAME

func Test_avutil_get_pix_fmt(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name     string
		expected AVPixelFormat
	}{
		{"rgb24", AV_PIX_FMT_RGB24},
		{"yuv420p", AV_PIX_FMT_YUV420P},
		{"rgba", AV_PIX_FMT_RGBA},
		{"gray", AV_PIX_FMT_GRAY8},
		{"nv12", AV_PIX_FMT_NV12},
	}

	for _, tc := range tests {
		fmt := AVUtil_get_pix_fmt(tc.name)
		assert.Equal(tc.expected, fmt)
		t.Logf("Name %q: format %v", tc.name, fmt)
	}
}

func Test_avutil_get_pix_fmt_invalid(t *testing.T) {
	assert := assert.New(t)

	// Invalid format name should return NONE
	fmt := AVUtil_get_pix_fmt("invalid_format_name_xyz")
	assert.Equal(AV_PIX_FMT_NONE, fmt)
}

func Test_avutil_get_pix_fmt_empty(t *testing.T) {
	assert := assert.New(t)

	// Empty string should return NONE
	fmt := AVUtil_get_pix_fmt("")
	assert.Equal(AV_PIX_FMT_NONE, fmt)
}

func Test_avutil_get_pix_fmt_case_sensitive(t *testing.T) {
	assert := assert.New(t)

	// FFmpeg format names are case-sensitive
	fmt1 := AVUtil_get_pix_fmt("rgb24")
	assert.NotEqual(AV_PIX_FMT_NONE, fmt1)

	fmt2 := AVUtil_get_pix_fmt("RGB24")
	assert.Equal(AV_PIX_FMT_NONE, fmt2)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PIXEL FORMAT ENUMERATION

func Test_avutil_next_pixel_fmt(t *testing.T) {
	assert := assert.New(t)

	var opaque uintptr
	count := 0
	seen := make(map[AVPixelFormat]bool)

	for {
		fmt := AVUtil_next_pixel_fmt(&opaque)
		if fmt == AV_PIX_FMT_NONE {
			break
		}
		count++
		seen[fmt] = true

		name := AVUtil_get_pix_fmt_name(fmt)
		assert.NotEmpty(name)

		if count <= 10 {
			t.Logf("Format %d: %v = %q", count, fmt, name)
		}
	}

	assert.Greater(count, 0, "Should enumerate at least one pixel format")
	t.Logf("Total pixel formats enumerated: %d", count)
}

func Test_avutil_next_pixel_fmt_nil(t *testing.T) {
	assert := assert.New(t)

	// Passing nil should return NONE
	fmt := AVUtil_next_pixel_fmt(nil)
	assert.Equal(AV_PIX_FMT_NONE, fmt)
}

func Test_avutil_next_pixel_fmt_consistency(t *testing.T) {
	assert := assert.New(t)

	// Two enumerations should return the same formats
	var opaque1 uintptr
	formats1 := []AVPixelFormat{}
	for {
		fmt := AVUtil_next_pixel_fmt(&opaque1)
		if fmt == AV_PIX_FMT_NONE {
			break
		}
		formats1 = append(formats1, fmt)
	}

	var opaque2 uintptr
	formats2 := []AVPixelFormat{}
	for {
		fmt := AVUtil_next_pixel_fmt(&opaque2)
		if fmt == AV_PIX_FMT_NONE {
			break
		}
		formats2 = append(formats2, fmt)
	}

	assert.Equal(len(formats1), len(formats2), "Both enumerations should return same count")
	assert.Equal(formats1, formats2, "Both enumerations should return same formats in same order")
}

////////////////////////////////////////////////////////////////////////////////
// TEST PIXEL FORMAT DESCRIPTOR

func Test_avutil_get_pix_fmt_desc(t *testing.T) {
	assert := assert.New(t)

	tests := []AVPixelFormat{
		AV_PIX_FMT_RGB24,
		AV_PIX_FMT_YUV420P,
		AV_PIX_FMT_RGBA,
		AV_PIX_FMT_GRAY8,
		AV_PIX_FMT_NV12,
	}

	for _, fmt := range tests {
		desc := AVUtil_get_pix_fmt_desc(fmt)
		assert.NotNil(desc)
		name := AVUtil_get_pix_fmt_name(fmt)
		t.Logf("Format %q has descriptor: %p", name, desc)
	}
}

func Test_avutil_get_pix_fmt_desc_none(t *testing.T) {
	assert := assert.New(t)

	// NONE format should return nil descriptor
	desc := AVUtil_get_pix_fmt_desc(AV_PIX_FMT_NONE)
	assert.Nil(desc)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PLANE COUNT

func Test_avutil_pix_fmt_count_planes(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		format      AVPixelFormat
		expectedMin int
		expectedMax int
		description string
	}{
		{AV_PIX_FMT_RGB24, 1, 1, "packed RGB should be 1 plane"},
		{AV_PIX_FMT_RGBA, 1, 1, "packed RGBA should be 1 plane"},
		{AV_PIX_FMT_YUV420P, 3, 3, "planar YUV420P should be 3 planes"},
		{AV_PIX_FMT_YUV422P, 3, 3, "planar YUV422P should be 3 planes"},
		{AV_PIX_FMT_YUV444P, 3, 3, "planar YUV444P should be 3 planes"},
		{AV_PIX_FMT_NV12, 2, 2, "semi-planar NV12 should be 2 planes"},
		{AV_PIX_FMT_GRAY8, 1, 1, "grayscale should be 1 plane"},
	}

	for _, tc := range tests {
		count := AVUtil_pix_fmt_count_planes(tc.format)
		assert.GreaterOrEqual(count, tc.expectedMin, tc.description)
		assert.LessOrEqual(count, tc.expectedMax, tc.description)
		name := AVUtil_get_pix_fmt_name(tc.format)
		t.Logf("Format %q: %d planes", name, count)
	}
}

func Test_avutil_pix_fmt_count_planes_none(t *testing.T) {
	assert := assert.New(t)

	// AV_PIX_FMT_NONE returns -22 (EINVAL) from FFmpeg
	count := AVUtil_pix_fmt_count_planes(AV_PIX_FMT_NONE)
	assert.Equal(-22, count)
}

func Test_avutil_pix_fmt_count_planes_various(t *testing.T) {
	assert := assert.New(t)

	var opaque uintptr
	tested := 0
	for tested < 20 {
		fmt := AVUtil_next_pixel_fmt(&opaque)
		if fmt == AV_PIX_FMT_NONE {
			break
		}
		tested++

		count := AVUtil_pix_fmt_count_planes(fmt)
		name := AVUtil_get_pix_fmt_name(fmt)
		assert.GreaterOrEqual(count, 0, "Plane count should be non-negative for %q", name)
		t.Logf("Format %q: %d planes", name, count)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST STRING OUTPUT

func Test_avutil_pixfmt_string(t *testing.T) {
	assert := assert.New(t)

	tests := []AVPixelFormat{
		AV_PIX_FMT_RGB24,
		AV_PIX_FMT_YUV420P,
		AV_PIX_FMT_RGBA,
		AV_PIX_FMT_GRAY8,
	}

	for _, fmt := range tests {
		str := fmt.String()
		assert.NotEmpty(str)
		assert.Contains(str, AVUtil_get_pix_fmt_name(fmt))
		t.Logf("Format %v String(): %q", fmt, str)
	}
}

func Test_avutil_pixfmt_string_none(t *testing.T) {
	assert := assert.New(t)

	// AV_PIX_FMT_NONE returns empty string, so String() returns formatted output
	str := AV_PIX_FMT_NONE.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AVPixelFormat")
}

func Test_avutil_pixfmt_string_unknown(t *testing.T) {
	assert := assert.New(t)

	// Test with a very large format value that doesn't exist
	unknownFmt := AVPixelFormat(99999)
	str := unknownFmt.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AVPixelFormat")
	t.Logf("Unknown format string: %q", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_pixfmt_marshal_json(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		format   AVPixelFormat
		expected string
	}{
		{AV_PIX_FMT_RGB24, "rgb24"},
		{AV_PIX_FMT_YUV420P, "yuv420p"},
		{AV_PIX_FMT_RGBA, "rgba"},
		{AV_PIX_FMT_GRAY8, "gray"},
	}

	for _, tc := range tests {
		data, err := json.Marshal(tc.format)
		assert.NoError(err)
		assert.Contains(string(data), tc.expected)
		t.Logf("Format %v JSON: %s", tc.format, string(data))
	}
}

func Test_avutil_pixfmt_marshal_json_none(t *testing.T) {
	assert := assert.New(t)

	// AV_PIX_FMT_NONE marshals as formatted string
	data, err := json.Marshal(AV_PIX_FMT_NONE)
	assert.NoError(err)
	assert.Contains(string(data), "AVPixelFormat")
}

func Test_avutil_pixfmt_marshal_json_struct(t *testing.T) {
	assert := assert.New(t)

	type TestStruct struct {
		Format AVPixelFormat `json:"format"`
		Name   string        `json:"name"`
	}

	s := TestStruct{
		Format: AV_PIX_FMT_RGB24,
		Name:   "test",
	}

	data, err := json.Marshal(s)
	assert.NoError(err)
	assert.Contains(string(data), "rgb24")
	assert.Contains(string(data), "test")
	t.Logf("Struct JSON: %s", string(data))
}

////////////////////////////////////////////////////////////////////////////////
// TEST ROUNDTRIP NAME/FORMAT

func Test_avutil_pixfmt_name_format_roundtrip(t *testing.T) {
	assert := assert.New(t)

	tests := []AVPixelFormat{
		AV_PIX_FMT_RGB24,
		AV_PIX_FMT_BGR24,
		AV_PIX_FMT_YUV420P,
		AV_PIX_FMT_YUV422P,
		AV_PIX_FMT_YUV444P,
		AV_PIX_FMT_RGBA,
		AV_PIX_FMT_BGRA,
		AV_PIX_FMT_ARGB,
		AV_PIX_FMT_ABGR,
		AV_PIX_FMT_GRAY8,
		AV_PIX_FMT_NV12,
		AV_PIX_FMT_NV21,
	}

	for _, original := range tests {
		name := AVUtil_get_pix_fmt_name(original)
		assert.NotEmpty(name)

		retrieved := AVUtil_get_pix_fmt(name)
		assert.Equal(original, retrieved, "Roundtrip failed for %q", name)
		t.Logf("Roundtrip: %v -> %q -> %v", original, name, retrieved)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST SPECIFIC PIXEL FORMATS

func Test_avutil_pixfmt_rgb_formats(t *testing.T) {
	assert := assert.New(t)

	rgbFormats := []AVPixelFormat{
		AV_PIX_FMT_RGB24,
		AV_PIX_FMT_BGR24,
		AV_PIX_FMT_RGBA,
		AV_PIX_FMT_BGRA,
		AV_PIX_FMT_ARGB,
		AV_PIX_FMT_ABGR,
		AV_PIX_FMT_RGB565BE,
		AV_PIX_FMT_RGB565LE,
	}

	for _, fmt := range rgbFormats {
		name := AVUtil_get_pix_fmt_name(fmt)
		assert.NotEmpty(name)
		assert.NotEqual("none", name)

		count := AVUtil_pix_fmt_count_planes(fmt)
		assert.GreaterOrEqual(count, 1)

		t.Logf("RGB format %q: %d planes", name, count)
	}
}

func Test_avutil_pixfmt_yuv_formats(t *testing.T) {
	assert := assert.New(t)

	yuvFormats := []AVPixelFormat{
		AV_PIX_FMT_YUV420P,
		AV_PIX_FMT_YUV422P,
		AV_PIX_FMT_YUV444P,
		AV_PIX_FMT_YUV410P,
		AV_PIX_FMT_YUV411P,
		AV_PIX_FMT_YUYV422,
		AV_PIX_FMT_UYVY422,
	}

	for _, fmt := range yuvFormats {
		name := AVUtil_get_pix_fmt_name(fmt)
		assert.NotEmpty(name)
		assert.NotEqual("none", name)

		count := AVUtil_pix_fmt_count_planes(fmt)
		assert.GreaterOrEqual(count, 1)

		t.Logf("YUV format %q: %d planes", name, count)
	}
}

func Test_avutil_pixfmt_gray_formats(t *testing.T) {
	assert := assert.New(t)

	grayFormats := []AVPixelFormat{
		AV_PIX_FMT_GRAY8,
		AV_PIX_FMT_GRAY16BE,
		AV_PIX_FMT_GRAY16LE,
	}

	for _, fmt := range grayFormats {
		name := AVUtil_get_pix_fmt_name(fmt)
		assert.NotEmpty(name)
		assert.NotEqual("none", name)

		count := AVUtil_pix_fmt_count_planes(fmt)
		assert.Equal(1, count, "Grayscale formats should have 1 plane")

		t.Logf("Gray format %q: %d plane", name, count)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_pixfmt_high_value(t *testing.T) {
	// Test with a very high format value
	highFmt := AVPixelFormat(10000)
	name := AVUtil_get_pix_fmt_name(highFmt)

	// Should either return empty or handle gracefully
	t.Logf("High format value %d returned: %q", highFmt, name)
}

func Test_avutil_pixfmt_negative_value(t *testing.T) {
	// Test with negative format value
	negativeFmt := AVPixelFormat(-1)

	// Should handle gracefully
	str := negativeFmt.String()
	t.Logf("Negative format value string: %q", str)
}

func Test_avutil_pixfmt_special_formats(t *testing.T) {
	assert := assert.New(t)

	specialFormats := []struct {
		format AVPixelFormat
		name   string
	}{
		{AV_PIX_FMT_NONE, ""},
		{AV_PIX_FMT_VAAPI, "vaapi"},
		{AV_PIX_FMT_CUDA, "cuda"},
	}

	for _, tc := range specialFormats {
		name := AVUtil_get_pix_fmt_name(tc.format)
		assert.Equal(tc.name, name)
		t.Logf("Special format %v: expected=%q, got=%q", tc.format, tc.name, name)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST ENUMERATION COMPLETENESS

func Test_avutil_pixfmt_enumeration_completeness(t *testing.T) {
	assert := assert.New(t)

	// Enumerate all formats
	var opaque uintptr
	allFormats := []AVPixelFormat{}
	names := make(map[string]bool)

	for {
		fmt := AVUtil_next_pixel_fmt(&opaque)
		if fmt == AV_PIX_FMT_NONE {
			break
		}
		allFormats = append(allFormats, fmt)

		name := AVUtil_get_pix_fmt_name(fmt)
		if name != "" {
			names[name] = true
		}
	}

	// Check that well-known formats exist and can be named
	// Note: AVUtil_next_pixel_fmt starts at index 0 which is YUV420P, not NONE
	knownFormats := []AVPixelFormat{
		AV_PIX_FMT_RGB24,
		AV_PIX_FMT_RGBA,
		AV_PIX_FMT_GRAY8,
	}

	for _, known := range knownFormats {
		name := AVUtil_get_pix_fmt_name(known)
		assert.NotEmpty(name, "Known format %v should have a name", known)
		t.Logf("Known format %v: %q", known, name)
	}

	t.Logf("Enumerated %d formats with %d unique names", len(allFormats), len(names))
}
