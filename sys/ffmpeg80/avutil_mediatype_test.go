package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST STRING OUTPUT

func Test_avutil_mediatype_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		mediaType AVMediaType
		expected  string
	}{
		{AVMEDIA_TYPE_UNKNOWN, "AVMEDIA_TYPE_UNKNOWN"},
		{AVMEDIA_TYPE_VIDEO, "AVMEDIA_TYPE_VIDEO"},
		{AVMEDIA_TYPE_AUDIO, "AVMEDIA_TYPE_AUDIO"},
		{AVMEDIA_TYPE_DATA, "AVMEDIA_TYPE_DATA"},
		{AVMEDIA_TYPE_SUBTITLE, "AVMEDIA_TYPE_SUBTITLE"},
		{AVMEDIA_TYPE_ATTACHMENT, "AVMEDIA_TYPE_ATTACHMENT"},
	}

	for _, tc := range tests {
		str := tc.mediaType.String()
		assert.Equal(tc.expected, str)
		t.Logf("MediaType %v: %q", tc.mediaType, str)
	}
}

func Test_avutil_mediatype_string_invalid(t *testing.T) {
	assert := assert.New(t)

	// Test with invalid media type value
	invalidType := AVMediaType(99)
	str := invalidType.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AVMediaType")
	assert.Contains(str, "99")
	t.Logf("Invalid media type string: %q", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_mediatype_json_marshaling(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		mediaType AVMediaType
		expected  string
	}{
		{AVMEDIA_TYPE_UNKNOWN, `"AVMEDIA_TYPE_UNKNOWN"`},
		{AVMEDIA_TYPE_VIDEO, `"AVMEDIA_TYPE_VIDEO"`},
		{AVMEDIA_TYPE_AUDIO, `"AVMEDIA_TYPE_AUDIO"`},
		{AVMEDIA_TYPE_DATA, `"AVMEDIA_TYPE_DATA"`},
		{AVMEDIA_TYPE_SUBTITLE, `"AVMEDIA_TYPE_SUBTITLE"`},
		{AVMEDIA_TYPE_ATTACHMENT, `"AVMEDIA_TYPE_ATTACHMENT"`},
	}

	for _, tc := range tests {
		jsonBytes, err := json.Marshal(tc.mediaType)
		assert.NoError(err)
		assert.Equal(tc.expected, string(jsonBytes))
		t.Logf("MediaType %v marshals to: %s", tc.mediaType, string(jsonBytes))
	}
}

func Test_avutil_mediatype_json_in_struct(t *testing.T) {
	assert := assert.New(t)

	type TestStruct struct {
		Type AVMediaType `json:"type"`
		Name string      `json:"name"`
	}

	testCases := []struct {
		name     string
		input    TestStruct
		expected string
	}{
		{
			name:     "video",
			input:    TestStruct{Type: AVMEDIA_TYPE_VIDEO, Name: "test_video"},
			expected: `{"type":"AVMEDIA_TYPE_VIDEO","name":"test_video"}`,
		},
		{
			name:     "audio",
			input:    TestStruct{Type: AVMEDIA_TYPE_AUDIO, Name: "test_audio"},
			expected: `{"type":"AVMEDIA_TYPE_AUDIO","name":"test_audio"}`,
		},
		{
			name:     "subtitle",
			input:    TestStruct{Type: AVMEDIA_TYPE_SUBTITLE, Name: "test_subtitle"},
			expected: `{"type":"AVMEDIA_TYPE_SUBTITLE","name":"test_subtitle"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tc.input)
			assert.NoError(err)
			assert.Equal(tc.expected, string(jsonBytes))
			t.Logf("Struct with %s marshals to: %s", tc.name, string(jsonBytes))
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST IS METHOD

func Test_avutil_mediatype_is_method(t *testing.T) {
	assert := assert.New(t)

	// Test positive matches
	assert.True(AVMEDIA_TYPE_VIDEO.Is(AVMEDIA_TYPE_VIDEO))
	assert.True(AVMEDIA_TYPE_AUDIO.Is(AVMEDIA_TYPE_AUDIO))
	assert.True(AVMEDIA_TYPE_SUBTITLE.Is(AVMEDIA_TYPE_SUBTITLE))
	assert.True(AVMEDIA_TYPE_DATA.Is(AVMEDIA_TYPE_DATA))
	assert.True(AVMEDIA_TYPE_ATTACHMENT.Is(AVMEDIA_TYPE_ATTACHMENT))
	assert.True(AVMEDIA_TYPE_UNKNOWN.Is(AVMEDIA_TYPE_UNKNOWN))

	// Test negative matches
	assert.False(AVMEDIA_TYPE_VIDEO.Is(AVMEDIA_TYPE_AUDIO))
	assert.False(AVMEDIA_TYPE_AUDIO.Is(AVMEDIA_TYPE_VIDEO))
	assert.False(AVMEDIA_TYPE_VIDEO.Is(AVMEDIA_TYPE_SUBTITLE))
	assert.False(AVMEDIA_TYPE_AUDIO.Is(AVMEDIA_TYPE_DATA))
	assert.False(AVMEDIA_TYPE_SUBTITLE.Is(AVMEDIA_TYPE_ATTACHMENT))
	assert.False(AVMEDIA_TYPE_DATA.Is(AVMEDIA_TYPE_UNKNOWN))
}

func Test_avutil_mediatype_is_with_variable(t *testing.T) {
	assert := assert.New(t)

	mediaType := AVMEDIA_TYPE_VIDEO

	// Test using Is method to check type
	if mediaType.Is(AVMEDIA_TYPE_VIDEO) {
		t.Log("Successfully identified as VIDEO type")
	} else {
		t.Error("Failed to identify VIDEO type")
	}

	// Test common pattern: checking if a type is video or audio
	isStreamType := mediaType.Is(AVMEDIA_TYPE_VIDEO) || mediaType.Is(AVMEDIA_TYPE_AUDIO)
	assert.True(isStreamType)

	// Test that non-stream types return false
	dataType := AVMEDIA_TYPE_DATA
	isStreamType = dataType.Is(AVMEDIA_TYPE_VIDEO) || dataType.Is(AVMEDIA_TYPE_AUDIO)
	assert.False(isStreamType)
}

////////////////////////////////////////////////////////////////////////////////
// TEST CONSTANTS UNIQUENESS

func Test_avutil_mediatype_constants_unique(t *testing.T) {
	assert := assert.New(t)

	types := []AVMediaType{
		AVMEDIA_TYPE_UNKNOWN,
		AVMEDIA_TYPE_VIDEO,
		AVMEDIA_TYPE_AUDIO,
		AVMEDIA_TYPE_DATA,
		AVMEDIA_TYPE_SUBTITLE,
		AVMEDIA_TYPE_ATTACHMENT,
	}

	// Check that all constants have unique values
	seen := make(map[AVMediaType]bool)
	for _, mediaType := range types {
		assert.False(seen[mediaType], "MediaType %v appears multiple times", mediaType)
		seen[mediaType] = true
		t.Logf("MediaType %v = %d", mediaType, int(mediaType))
	}

	// Verify we have exactly 6 unique media types
	assert.Equal(6, len(seen))
}

func Test_avutil_mediatype_constants_values(t *testing.T) {
	assert := assert.New(t)

	// Verify that the Go constants match the expected C values
	// AVMEDIA_TYPE_UNKNOWN should be -1
	// AVMEDIA_TYPE_VIDEO should be 0
	// These values are defined by FFmpeg

	t.Logf("AVMEDIA_TYPE_UNKNOWN = %d", int(AVMEDIA_TYPE_UNKNOWN))
	t.Logf("AVMEDIA_TYPE_VIDEO = %d", int(AVMEDIA_TYPE_VIDEO))
	t.Logf("AVMEDIA_TYPE_AUDIO = %d", int(AVMEDIA_TYPE_AUDIO))
	t.Logf("AVMEDIA_TYPE_DATA = %d", int(AVMEDIA_TYPE_DATA))
	t.Logf("AVMEDIA_TYPE_SUBTITLE = %d", int(AVMEDIA_TYPE_SUBTITLE))
	t.Logf("AVMEDIA_TYPE_ATTACHMENT = %d", int(AVMEDIA_TYPE_ATTACHMENT))

	// Ensure VIDEO comes before AUDIO (standard FFmpeg ordering)
	assert.True(int(AVMEDIA_TYPE_VIDEO) < int(AVMEDIA_TYPE_AUDIO))
}

////////////////////////////////////////////////////////////////////////////////
// TEST COMMON USE CASES

func Test_avutil_mediatype_switch_statement(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		mediaType    AVMediaType
		expectedDesc string
	}{
		{AVMEDIA_TYPE_VIDEO, "video stream"},
		{AVMEDIA_TYPE_AUDIO, "audio stream"},
		{AVMEDIA_TYPE_SUBTITLE, "subtitle stream"},
		{AVMEDIA_TYPE_DATA, "data stream"},
		{AVMEDIA_TYPE_ATTACHMENT, "attachment"},
		{AVMEDIA_TYPE_UNKNOWN, "unknown"},
	}

	for _, tc := range testCases {
		var description string
		switch tc.mediaType {
		case AVMEDIA_TYPE_VIDEO:
			description = "video stream"
		case AVMEDIA_TYPE_AUDIO:
			description = "audio stream"
		case AVMEDIA_TYPE_SUBTITLE:
			description = "subtitle stream"
		case AVMEDIA_TYPE_DATA:
			description = "data stream"
		case AVMEDIA_TYPE_ATTACHMENT:
			description = "attachment"
		case AVMEDIA_TYPE_UNKNOWN:
			description = "unknown"
		default:
			description = "invalid"
		}

		assert.Equal(tc.expectedDesc, description)
		t.Logf("%s -> %s", tc.mediaType.String(), description)
	}
}

func Test_avutil_mediatype_array_usage(t *testing.T) {
	assert := assert.New(t)

	// Common pattern: array or slice of media types
	supportedTypes := []AVMediaType{
		AVMEDIA_TYPE_VIDEO,
		AVMEDIA_TYPE_AUDIO,
	}

	// Check if a type is supported
	checkType := AVMEDIA_TYPE_VIDEO
	found := false
	for _, supportedType := range supportedTypes {
		if checkType.Is(supportedType) {
			found = true
			break
		}
	}
	assert.True(found)

	// Check unsupported type
	checkType = AVMEDIA_TYPE_SUBTITLE
	found = false
	for _, supportedType := range supportedTypes {
		if checkType.Is(supportedType) {
			found = true
			break
		}
	}
	assert.False(found)
}

func Test_avutil_mediatype_map_usage(t *testing.T) {
	assert := assert.New(t)

	// Common pattern: map with media type as key
	streamCounts := map[AVMediaType]int{
		AVMEDIA_TYPE_VIDEO:    1,
		AVMEDIA_TYPE_AUDIO:    2,
		AVMEDIA_TYPE_SUBTITLE: 3,
	}

	assert.Equal(1, streamCounts[AVMEDIA_TYPE_VIDEO])
	assert.Equal(2, streamCounts[AVMEDIA_TYPE_AUDIO])
	assert.Equal(3, streamCounts[AVMEDIA_TYPE_SUBTITLE])
	assert.Equal(0, streamCounts[AVMEDIA_TYPE_DATA]) // Not in map

	// Verify all keys are valid
	for mediaType, count := range streamCounts {
		assert.NotEmpty(mediaType.String())
		assert.Greater(count, 0)
		t.Logf("StreamCount[%s] = %d", mediaType.String(), count)
	}
}
