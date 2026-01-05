package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST STRING OUTPUT

func Test_avutil_picturetype_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		pictType AVPictureType
		expected string
	}{
		{AV_PICTURE_TYPE_NONE, "NONE"},
		{AV_PICTURE_TYPE_I, "I"},
		{AV_PICTURE_TYPE_P, "P"},
		{AV_PICTURE_TYPE_B, "B"},
		{AV_PICTURE_TYPE_S, "S"},
		{AV_PICTURE_TYPE_SI, "SI"},
		{AV_PICTURE_TYPE_SP, "SP"},
		{AV_PICTURE_TYPE_BI, "BI"},
	}

	for _, tc := range tests {
		str := tc.pictType.String()
		assert.Equal(tc.expected, str)
		t.Logf("PictureType %v: %q", tc.pictType, str)
	}
}

func Test_avutil_picturetype_string_invalid(t *testing.T) {
	assert := assert.New(t)

	// Test with invalid picture type value
	invalidType := AVPictureType(99)
	str := invalidType.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AVPictureType")
	t.Logf("Invalid picture type string: %q", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_picturetype_marshal_json(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		pictType AVPictureType
		expected string
	}{
		{AV_PICTURE_TYPE_NONE, "NONE"},
		{AV_PICTURE_TYPE_I, "I"},
		{AV_PICTURE_TYPE_P, "P"},
		{AV_PICTURE_TYPE_B, "B"},
		{AV_PICTURE_TYPE_S, "S"},
	}

	for _, tc := range tests {
		data, err := json.Marshal(tc.pictType)
		assert.NoError(err)
		assert.Contains(string(data), tc.expected)
		t.Logf("PictureType %v JSON: %s", tc.pictType, string(data))
	}
}

func Test_avutil_picturetype_marshal_json_struct(t *testing.T) {
	assert := assert.New(t)

	type TestStruct struct {
		Type AVPictureType `json:"type"`
		Name string        `json:"name"`
	}

	s := TestStruct{
		Type: AV_PICTURE_TYPE_I,
		Name: "keyframe",
	}

	data, err := json.Marshal(s)
	assert.NoError(err)
	assert.Contains(string(data), "I")
	assert.Contains(string(data), "keyframe")
	t.Logf("Struct JSON: %s", string(data))
}

////////////////////////////////////////////////////////////////////////////////
// TEST PICTURE TYPE CHARACTER

func Test_avutil_get_picture_type_char(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		pictType     AVPictureType
		expectedChar rune
	}{
		{AV_PICTURE_TYPE_NONE, '?'},
		{AV_PICTURE_TYPE_I, 'I'},
		{AV_PICTURE_TYPE_P, 'P'},
		{AV_PICTURE_TYPE_B, 'B'},
		{AV_PICTURE_TYPE_S, 'S'},
		{AV_PICTURE_TYPE_SI, 'i'},
		{AV_PICTURE_TYPE_SP, 'p'},
		{AV_PICTURE_TYPE_BI, 'b'},
	}

	for _, tc := range tests {
		char := AVUtil_get_picture_type_char(tc.pictType)
		assert.Equal(tc.expectedChar, char)
		t.Logf("PictureType %v: char=%c", tc.pictType, char)
	}
}

func Test_avutil_get_picture_type_char_all_types(t *testing.T) {
	types := []AVPictureType{
		AV_PICTURE_TYPE_NONE,
		AV_PICTURE_TYPE_I,
		AV_PICTURE_TYPE_P,
		AV_PICTURE_TYPE_B,
		AV_PICTURE_TYPE_S,
		AV_PICTURE_TYPE_SI,
		AV_PICTURE_TYPE_SP,
		AV_PICTURE_TYPE_BI,
	}

	for _, pt := range types {
		char := AVUtil_get_picture_type_char(pt)
		t.Logf("PictureType %s -> '%c'", pt.String(), char)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST ALL PICTURE TYPES

func Test_avutil_picturetype_all_defined(t *testing.T) {
	assert := assert.New(t)

	types := []AVPictureType{
		AV_PICTURE_TYPE_NONE,
		AV_PICTURE_TYPE_I,
		AV_PICTURE_TYPE_P,
		AV_PICTURE_TYPE_B,
		AV_PICTURE_TYPE_S,
		AV_PICTURE_TYPE_SI,
		AV_PICTURE_TYPE_SP,
		AV_PICTURE_TYPE_BI,
	}

	for _, pt := range types {
		// Each type should have a string representation
		str := pt.String()
		assert.NotEmpty(str)

		// Each type should have a character representation
		char := AVUtil_get_picture_type_char(pt)
		assert.NotEqual(rune(0), char)

		// Each type should be JSON marshalable
		data, err := json.Marshal(pt)
		assert.NoError(err)
		assert.NotEmpty(data)

		t.Logf("PictureType %v: string=%q, char=%c", pt, str, char)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST COMMON PICTURE TYPES

func Test_avutil_picturetype_common_types(t *testing.T) {
	assert := assert.New(t)

	// I-frame (keyframe)
	iType := AV_PICTURE_TYPE_I
	assert.Equal("I", iType.String())
	assert.Equal('I', AVUtil_get_picture_type_char(iType))

	// P-frame (predicted)
	pType := AV_PICTURE_TYPE_P
	assert.Equal("P", pType.String())
	assert.Equal('P', AVUtil_get_picture_type_char(pType))

	// B-frame (bi-directional)
	bType := AV_PICTURE_TYPE_B
	assert.Equal("B", bType.String())
	assert.Equal('B', AVUtil_get_picture_type_char(bType))
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_picturetype_constants_unique(t *testing.T) {
	assert := assert.New(t)

	// All constants should have unique values
	types := []AVPictureType{
		AV_PICTURE_TYPE_NONE,
		AV_PICTURE_TYPE_I,
		AV_PICTURE_TYPE_P,
		AV_PICTURE_TYPE_B,
		AV_PICTURE_TYPE_S,
		AV_PICTURE_TYPE_SI,
		AV_PICTURE_TYPE_SP,
		AV_PICTURE_TYPE_BI,
	}

	seen := make(map[AVPictureType]bool)
	for _, pt := range types {
		assert.False(seen[pt], "Duplicate picture type value: %v", pt)
		seen[pt] = true
	}

	assert.Equal(len(types), len(seen), "All picture types should be unique")
}

func Test_avutil_picturetype_zero_value(t *testing.T) {
	assert := assert.New(t)

	// Zero value should be valid
	var pt AVPictureType
	str := pt.String()
	assert.NotEmpty(str)
	t.Logf("Zero value PictureType: %q", str)
}

func Test_avutil_picturetype_negative_value(t *testing.T) {
	// Test with very large value that doesn't exist
	largePt := AVPictureType(100)
	str := largePt.String()
	t.Logf("Large picture type value string: %q", str)
}
