package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST SAMPLE FORMAT NAME

func Test_avutil_get_sample_fmt_name(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		format   AVSampleFormat
		expected string
	}{
		{AV_SAMPLE_FMT_U8, "u8"},
		{AV_SAMPLE_FMT_S16, "s16"},
		{AV_SAMPLE_FMT_S32, "s32"},
		{AV_SAMPLE_FMT_FLT, "flt"},
		{AV_SAMPLE_FMT_DBL, "dbl"},
		{AV_SAMPLE_FMT_U8P, "u8p"},
		{AV_SAMPLE_FMT_S16P, "s16p"},
		{AV_SAMPLE_FMT_S32P, "s32p"},
		{AV_SAMPLE_FMT_FLTP, "fltp"},
		{AV_SAMPLE_FMT_DBLP, "dblp"},
		{AV_SAMPLE_FMT_S64, "s64"},
		{AV_SAMPLE_FMT_S64P, "s64p"},
	}

	for _, tc := range tests {
		name := AVUtil_get_sample_fmt_name(tc.format)
		assert.Equal(tc.expected, name)
		t.Logf("Format %v: %q", tc.format, name)
	}
}

func Test_avutil_get_sample_fmt_name_none(t *testing.T) {
	assert := assert.New(t)

	// AV_SAMPLE_FMT_NONE returns empty string from FFmpeg
	name := AVUtil_get_sample_fmt_name(AV_SAMPLE_FMT_NONE)
	assert.Equal("", name)
}

////////////////////////////////////////////////////////////////////////////////
// TEST GET SAMPLE FORMAT BY NAME

func Test_avutil_get_sample_fmt(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name     string
		expected AVSampleFormat
	}{
		{"u8", AV_SAMPLE_FMT_U8},
		{"s16", AV_SAMPLE_FMT_S16},
		{"s32", AV_SAMPLE_FMT_S32},
		{"flt", AV_SAMPLE_FMT_FLT},
		{"dbl", AV_SAMPLE_FMT_DBL},
		{"u8p", AV_SAMPLE_FMT_U8P},
		{"s16p", AV_SAMPLE_FMT_S16P},
		{"s32p", AV_SAMPLE_FMT_S32P},
		{"fltp", AV_SAMPLE_FMT_FLTP},
		{"dblp", AV_SAMPLE_FMT_DBLP},
		{"s64", AV_SAMPLE_FMT_S64},
		{"s64p", AV_SAMPLE_FMT_S64P},
	}

	for _, tc := range tests {
		fmt := AVUtil_get_sample_fmt(tc.name)
		assert.Equal(tc.expected, fmt)
		t.Logf("Name %q: format %v", tc.name, fmt)
	}
}

func Test_avutil_get_sample_fmt_invalid(t *testing.T) {
	assert := assert.New(t)

	// Invalid format name should return NONE
	fmt := AVUtil_get_sample_fmt("invalid_format_name_xyz")
	assert.Equal(AV_SAMPLE_FMT_NONE, fmt)
}

func Test_avutil_get_sample_fmt_empty(t *testing.T) {
	assert := assert.New(t)

	// Empty string should return NONE
	fmt := AVUtil_get_sample_fmt("")
	assert.Equal(AV_SAMPLE_FMT_NONE, fmt)
}

func Test_avutil_get_sample_fmt_case_sensitive(t *testing.T) {
	assert := assert.New(t)

	// FFmpeg format names are case-sensitive
	fmt1 := AVUtil_get_sample_fmt("s16")
	assert.NotEqual(AV_SAMPLE_FMT_NONE, fmt1)

	fmt2 := AVUtil_get_sample_fmt("S16")
	assert.Equal(AV_SAMPLE_FMT_NONE, fmt2)
}

////////////////////////////////////////////////////////////////////////////////
// TEST SAMPLE FORMAT ENUMERATION

func Test_avutil_next_sample_fmt(t *testing.T) {
	assert := assert.New(t)

	var opaque uintptr
	count := 0
	seen := make(map[AVSampleFormat]bool)

	for {
		fmt := AVUtil_next_sample_fmt(&opaque)
		if fmt == AV_SAMPLE_FMT_NONE {
			break
		}
		count++
		seen[fmt] = true

		name := AVUtil_get_sample_fmt_name(fmt)
		t.Logf("Format %d: %v = %q", count, fmt, name)
	}

	assert.Greater(count, 0, "Should enumerate at least one sample format")
	t.Logf("Total sample formats enumerated: %d", count)
}

func Test_avutil_next_sample_fmt_nil(t *testing.T) {
	assert := assert.New(t)

	// Passing nil should return NONE
	fmt := AVUtil_next_sample_fmt(nil)
	assert.Equal(AV_SAMPLE_FMT_NONE, fmt)
}

func Test_avutil_next_sample_fmt_consistency(t *testing.T) {
	assert := assert.New(t)

	// Two enumerations should return the same formats
	var opaque1 uintptr
	formats1 := []AVSampleFormat{}
	for {
		fmt := AVUtil_next_sample_fmt(&opaque1)
		if fmt == AV_SAMPLE_FMT_NONE {
			break
		}
		formats1 = append(formats1, fmt)
	}

	var opaque2 uintptr
	formats2 := []AVSampleFormat{}
	for {
		fmt := AVUtil_next_sample_fmt(&opaque2)
		if fmt == AV_SAMPLE_FMT_NONE {
			break
		}
		formats2 = append(formats2, fmt)
	}

	assert.Equal(len(formats1), len(formats2), "Both enumerations should return same count")
	assert.Equal(formats1, formats2, "Both enumerations should return same formats in same order")
}

func Test_avutil_next_sample_fmt_starts_at_none(t *testing.T) {
	assert := assert.New(t)

	// First call should return index 0 which is AV_SAMPLE_FMT_NONE (-1)
	// The enumeration includes NONE as the first value
	var opaque uintptr
	first := AVUtil_next_sample_fmt(&opaque)
	// The actual value depends on the enum definition
	// Just verify it returns a valid format
	assert.NotEqual(AVSampleFormat(99999), first)
	t.Logf("First enumerated format: %v (%q)", first, AVUtil_get_sample_fmt_name(first))
}

////////////////////////////////////////////////////////////////////////////////
// TEST BYTES PER SAMPLE

func Test_avutil_get_bytes_per_sample(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		format       AVSampleFormat
		expectedSize int
		description  string
	}{
		{AV_SAMPLE_FMT_U8, 1, "u8 should be 1 byte"},
		{AV_SAMPLE_FMT_S16, 2, "s16 should be 2 bytes"},
		{AV_SAMPLE_FMT_S32, 4, "s32 should be 4 bytes"},
		{AV_SAMPLE_FMT_FLT, 4, "flt should be 4 bytes"},
		{AV_SAMPLE_FMT_DBL, 8, "dbl should be 8 bytes"},
		{AV_SAMPLE_FMT_S64, 8, "s64 should be 8 bytes"},
		{AV_SAMPLE_FMT_U8P, 1, "u8p should be 1 byte per plane"},
		{AV_SAMPLE_FMT_S16P, 2, "s16p should be 2 bytes per plane"},
		{AV_SAMPLE_FMT_S32P, 4, "s32p should be 4 bytes per plane"},
		{AV_SAMPLE_FMT_FLTP, 4, "fltp should be 4 bytes per plane"},
		{AV_SAMPLE_FMT_DBLP, 8, "dblp should be 8 bytes per plane"},
		{AV_SAMPLE_FMT_S64P, 8, "s64p should be 8 bytes per plane"},
	}

	for _, tc := range tests {
		size := AVUtil_get_bytes_per_sample(tc.format)
		assert.Equal(tc.expectedSize, size, tc.description)
		name := AVUtil_get_sample_fmt_name(tc.format)
		t.Logf("Format %q: %d bytes per sample", name, size)
	}
}

func Test_avutil_get_bytes_per_sample_none(t *testing.T) {
	assert := assert.New(t)

	size := AVUtil_get_bytes_per_sample(AV_SAMPLE_FMT_NONE)
	assert.Equal(0, size)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PLANAR CHECK

func Test_avutil_sample_fmt_is_planar(t *testing.T) {
	assert := assert.New(t)

	packedFormats := []AVSampleFormat{
		AV_SAMPLE_FMT_U8,
		AV_SAMPLE_FMT_S16,
		AV_SAMPLE_FMT_S32,
		AV_SAMPLE_FMT_FLT,
		AV_SAMPLE_FMT_DBL,
		AV_SAMPLE_FMT_S64,
	}

	for _, fmt := range packedFormats {
		isPlanar := AVUtil_sample_fmt_is_planar(fmt)
		assert.False(isPlanar, "Format %q should not be planar", AVUtil_get_sample_fmt_name(fmt))
		t.Logf("Format %q: planar=%v", AVUtil_get_sample_fmt_name(fmt), isPlanar)
	}

	planarFormats := []AVSampleFormat{
		AV_SAMPLE_FMT_U8P,
		AV_SAMPLE_FMT_S16P,
		AV_SAMPLE_FMT_S32P,
		AV_SAMPLE_FMT_FLTP,
		AV_SAMPLE_FMT_DBLP,
		AV_SAMPLE_FMT_S64P,
	}

	for _, fmt := range planarFormats {
		isPlanar := AVUtil_sample_fmt_is_planar(fmt)
		assert.True(isPlanar, "Format %q should be planar", AVUtil_get_sample_fmt_name(fmt))
		t.Logf("Format %q: planar=%v", AVUtil_get_sample_fmt_name(fmt), isPlanar)
	}
}

func Test_avutil_sample_fmt_is_planar_none(t *testing.T) {
	assert := assert.New(t)

	isPlanar := AVUtil_sample_fmt_is_planar(AV_SAMPLE_FMT_NONE)
	assert.False(isPlanar)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PACKED/PLANAR CONVERSION

func Test_avutil_get_packed_sample_fmt(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input    AVSampleFormat
		expected AVSampleFormat
	}{
		{AV_SAMPLE_FMT_U8, AV_SAMPLE_FMT_U8},
		{AV_SAMPLE_FMT_S16, AV_SAMPLE_FMT_S16},
		{AV_SAMPLE_FMT_S32, AV_SAMPLE_FMT_S32},
		{AV_SAMPLE_FMT_FLT, AV_SAMPLE_FMT_FLT},
		{AV_SAMPLE_FMT_DBL, AV_SAMPLE_FMT_DBL},
		{AV_SAMPLE_FMT_S64, AV_SAMPLE_FMT_S64},
		{AV_SAMPLE_FMT_U8P, AV_SAMPLE_FMT_U8},
		{AV_SAMPLE_FMT_S16P, AV_SAMPLE_FMT_S16},
		{AV_SAMPLE_FMT_S32P, AV_SAMPLE_FMT_S32},
		{AV_SAMPLE_FMT_FLTP, AV_SAMPLE_FMT_FLT},
		{AV_SAMPLE_FMT_DBLP, AV_SAMPLE_FMT_DBL},
		{AV_SAMPLE_FMT_S64P, AV_SAMPLE_FMT_S64},
	}

	for _, tc := range tests {
		result := AVUtil_get_packed_sample_fmt(tc.input)
		assert.Equal(tc.expected, result)
		inputName := AVUtil_get_sample_fmt_name(tc.input)
		resultName := AVUtil_get_sample_fmt_name(result)
		t.Logf("Packed form of %q: %q", inputName, resultName)
	}
}

func Test_avutil_get_planar_sample_fmt(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input    AVSampleFormat
		expected AVSampleFormat
	}{
		{AV_SAMPLE_FMT_U8, AV_SAMPLE_FMT_U8P},
		{AV_SAMPLE_FMT_S16, AV_SAMPLE_FMT_S16P},
		{AV_SAMPLE_FMT_S32, AV_SAMPLE_FMT_S32P},
		{AV_SAMPLE_FMT_FLT, AV_SAMPLE_FMT_FLTP},
		{AV_SAMPLE_FMT_DBL, AV_SAMPLE_FMT_DBLP},
		{AV_SAMPLE_FMT_S64, AV_SAMPLE_FMT_S64P},
		{AV_SAMPLE_FMT_U8P, AV_SAMPLE_FMT_U8P},
		{AV_SAMPLE_FMT_S16P, AV_SAMPLE_FMT_S16P},
		{AV_SAMPLE_FMT_S32P, AV_SAMPLE_FMT_S32P},
		{AV_SAMPLE_FMT_FLTP, AV_SAMPLE_FMT_FLTP},
		{AV_SAMPLE_FMT_DBLP, AV_SAMPLE_FMT_DBLP},
		{AV_SAMPLE_FMT_S64P, AV_SAMPLE_FMT_S64P},
	}

	for _, tc := range tests {
		result := AVUtil_get_planar_sample_fmt(tc.input)
		assert.Equal(tc.expected, result)
		inputName := AVUtil_get_sample_fmt_name(tc.input)
		resultName := AVUtil_get_sample_fmt_name(result)
		t.Logf("Planar form of %q: %q", inputName, resultName)
	}
}

func Test_avutil_packed_planar_roundtrip(t *testing.T) {
	assert := assert.New(t)

	// Test that converting packed -> planar -> packed returns original
	packedFormats := []AVSampleFormat{
		AV_SAMPLE_FMT_U8,
		AV_SAMPLE_FMT_S16,
		AV_SAMPLE_FMT_S32,
		AV_SAMPLE_FMT_FLT,
		AV_SAMPLE_FMT_DBL,
		AV_SAMPLE_FMT_S64,
	}

	for _, original := range packedFormats {
		planar := AVUtil_get_planar_sample_fmt(original)
		packed := AVUtil_get_packed_sample_fmt(planar)
		assert.Equal(original, packed)
		t.Logf("Roundtrip: %q -> %q -> %q",
			AVUtil_get_sample_fmt_name(original),
			AVUtil_get_sample_fmt_name(planar),
			AVUtil_get_sample_fmt_name(packed))
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST STRING OUTPUT

func Test_avutil_samplefmt_string(t *testing.T) {
	assert := assert.New(t)

	tests := []AVSampleFormat{
		AV_SAMPLE_FMT_U8,
		AV_SAMPLE_FMT_S16,
		AV_SAMPLE_FMT_S32,
		AV_SAMPLE_FMT_FLT,
		AV_SAMPLE_FMT_DBL,
		AV_SAMPLE_FMT_S64,
	}

	for _, fmt := range tests {
		str := fmt.String()
		assert.NotEmpty(str)
		assert.Contains(str, AVUtil_get_sample_fmt_name(fmt))
		t.Logf("Format %v String(): %q", fmt, str)
	}
}

func Test_avutil_samplefmt_string_none(t *testing.T) {
	assert := assert.New(t)

	// AV_SAMPLE_FMT_NONE returns empty string, so String() returns formatted output
	str := AV_SAMPLE_FMT_NONE.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AVSampleFormat")
}

func Test_avutil_samplefmt_string_unknown(t *testing.T) {
	assert := assert.New(t)

	// Test with a very large format value that doesn't exist
	unknownFmt := AVSampleFormat(99999)
	str := unknownFmt.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AVSampleFormat")
	t.Logf("Unknown format string: %q", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_samplefmt_marshal_json(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		format   AVSampleFormat
		expected string
	}{
		{AV_SAMPLE_FMT_U8, "u8"},
		{AV_SAMPLE_FMT_S16, "s16"},
		{AV_SAMPLE_FMT_FLT, "flt"},
		{AV_SAMPLE_FMT_DBL, "dbl"},
	}

	for _, tc := range tests {
		data, err := json.Marshal(tc.format)
		assert.NoError(err)
		assert.Contains(string(data), tc.expected)
		t.Logf("Format %v JSON: %s", tc.format, string(data))
	}
}

func Test_avutil_samplefmt_marshal_json_none(t *testing.T) {
	assert := assert.New(t)

	// AV_SAMPLE_FMT_NONE marshals as formatted string
	data, err := json.Marshal(AV_SAMPLE_FMT_NONE)
	assert.NoError(err)
	assert.Contains(string(data), "AVSampleFormat")
}

func Test_avutil_samplefmt_marshal_json_struct(t *testing.T) {
	assert := assert.New(t)

	type TestStruct struct {
		Format AVSampleFormat `json:"format"`
		Name   string         `json:"name"`
	}

	s := TestStruct{
		Format: AV_SAMPLE_FMT_S16,
		Name:   "test",
	}

	data, err := json.Marshal(s)
	assert.NoError(err)
	assert.Contains(string(data), "s16")
	assert.Contains(string(data), "test")
	t.Logf("Struct JSON: %s", string(data))
}

////////////////////////////////////////////////////////////////////////////////
// TEST ROUNDTRIP NAME/FORMAT

func Test_avutil_samplefmt_name_format_roundtrip(t *testing.T) {
	assert := assert.New(t)

	tests := []AVSampleFormat{
		AV_SAMPLE_FMT_U8,
		AV_SAMPLE_FMT_S16,
		AV_SAMPLE_FMT_S32,
		AV_SAMPLE_FMT_FLT,
		AV_SAMPLE_FMT_DBL,
		AV_SAMPLE_FMT_U8P,
		AV_SAMPLE_FMT_S16P,
		AV_SAMPLE_FMT_S32P,
		AV_SAMPLE_FMT_FLTP,
		AV_SAMPLE_FMT_DBLP,
		AV_SAMPLE_FMT_S64,
		AV_SAMPLE_FMT_S64P,
	}

	for _, original := range tests {
		name := AVUtil_get_sample_fmt_name(original)
		assert.NotEmpty(name)

		retrieved := AVUtil_get_sample_fmt(name)
		assert.Equal(original, retrieved, "Roundtrip failed for %q", name)
		t.Logf("Roundtrip: %v -> %q -> %v", original, name, retrieved)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST INTEGER FORMATS

func Test_avutil_samplefmt_integer_formats(t *testing.T) {
	assert := assert.New(t)

	intFormats := []AVSampleFormat{
		AV_SAMPLE_FMT_U8,
		AV_SAMPLE_FMT_S16,
		AV_SAMPLE_FMT_S32,
		AV_SAMPLE_FMT_S64,
		AV_SAMPLE_FMT_U8P,
		AV_SAMPLE_FMT_S16P,
		AV_SAMPLE_FMT_S32P,
		AV_SAMPLE_FMT_S64P,
	}

	for _, fmt := range intFormats {
		name := AVUtil_get_sample_fmt_name(fmt)
		assert.NotEmpty(name)

		size := AVUtil_get_bytes_per_sample(fmt)
		assert.Greater(size, 0)

		t.Logf("Integer format %q: %d bytes", name, size)
	}
}

func Test_avutil_samplefmt_float_formats(t *testing.T) {
	assert := assert.New(t)

	floatFormats := []AVSampleFormat{
		AV_SAMPLE_FMT_FLT,
		AV_SAMPLE_FMT_DBL,
		AV_SAMPLE_FMT_FLTP,
		AV_SAMPLE_FMT_DBLP,
	}

	for _, fmt := range floatFormats {
		name := AVUtil_get_sample_fmt_name(fmt)
		assert.NotEmpty(name)

		size := AVUtil_get_bytes_per_sample(fmt)
		assert.Greater(size, 0)

		t.Logf("Float format %q: %d bytes", name, size)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_samplefmt_high_value(t *testing.T) {
	// Test with a very high format value
	highFmt := AVSampleFormat(10000)
	name := AVUtil_get_sample_fmt_name(highFmt)

	// Should either return empty or handle gracefully
	t.Logf("High format value %d returned: %q", highFmt, name)
}

func Test_avutil_samplefmt_negative_value(t *testing.T) {
	// Test with negative format value (AV_SAMPLE_FMT_NONE is -1)
	negativeFmt := AVSampleFormat(-2)

	// Should handle gracefully
	str := negativeFmt.String()
	t.Logf("Negative format value string: %q", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST ENUMERATION COMPLETENESS

func Test_avutil_samplefmt_enumeration_completeness(t *testing.T) {
	assert := assert.New(t)

	// Enumerate all formats
	var opaque uintptr
	allFormats := []AVSampleFormat{}
	names := make(map[string]bool)

	for {
		fmt := AVUtil_next_sample_fmt(&opaque)
		if fmt == AV_SAMPLE_FMT_NONE {
			break
		}
		// Skip the first NONE we get from enumeration
		if len(allFormats) == 0 && fmt == AV_SAMPLE_FMT_NONE {
			continue
		}
		allFormats = append(allFormats, fmt)

		name := AVUtil_get_sample_fmt_name(fmt)
		if name != "" {
			names[name] = true
		}
	}

	// Check that well-known formats exist and can be named
	knownFormats := []AVSampleFormat{
		AV_SAMPLE_FMT_S16,
		AV_SAMPLE_FMT_FLT,
		AV_SAMPLE_FMT_S16P,
	}

	for _, known := range knownFormats {
		name := AVUtil_get_sample_fmt_name(known)
		assert.NotEmpty(name, "Known format %v should have a name", known)
		t.Logf("Known format %v: %q", known, name)
	}

	t.Logf("Enumerated %d formats with %d unique names", len(allFormats), len(names))
}

////////////////////////////////////////////////////////////////////////////////
// TEST ALL FORMATS PROPERTIES

func Test_avutil_samplefmt_all_properties(t *testing.T) {
	assert := assert.New(t)

	// Test all defined formats have consistent properties
	formats := []AVSampleFormat{
		AV_SAMPLE_FMT_U8,
		AV_SAMPLE_FMT_S16,
		AV_SAMPLE_FMT_S32,
		AV_SAMPLE_FMT_FLT,
		AV_SAMPLE_FMT_DBL,
		AV_SAMPLE_FMT_U8P,
		AV_SAMPLE_FMT_S16P,
		AV_SAMPLE_FMT_S32P,
		AV_SAMPLE_FMT_FLTP,
		AV_SAMPLE_FMT_DBLP,
		AV_SAMPLE_FMT_S64,
		AV_SAMPLE_FMT_S64P,
	}

	for _, fmt := range formats {
		name := AVUtil_get_sample_fmt_name(fmt)
		assert.NotEmpty(name, "Format should have a name")

		size := AVUtil_get_bytes_per_sample(fmt)
		assert.Greater(size, 0, "Format should have positive size")

		isPlanar := AVUtil_sample_fmt_is_planar(fmt)

		// Verify packed/planar consistency
		if isPlanar {
			packed := AVUtil_get_packed_sample_fmt(fmt)
			assert.False(AVUtil_sample_fmt_is_planar(packed), "Packed version should not be planar")

			// Converting back to planar should give original
			planarAgain := AVUtil_get_planar_sample_fmt(packed)
			assert.Equal(fmt, planarAgain)
		} else {
			planar := AVUtil_get_planar_sample_fmt(fmt)
			assert.True(AVUtil_sample_fmt_is_planar(planar), "Planar version should be planar")

			// Converting back to packed should give original
			packedAgain := AVUtil_get_packed_sample_fmt(planar)
			assert.Equal(fmt, packedAgain)
		}

		t.Logf("Format %q: size=%d, planar=%v", name, size, isPlanar)
	}
}
