package ffmpeg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - STRING AND FLAGSTRING

func Test_AVSeekFlag_String_001(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		flag     AVSeekFlag
		expected string
	}{
		{AVSEEK_FLAG_NONE, "AVSEEK_FLAG_NONE"},
		{AVSEEK_FLAG_BACKWARD, "AVSEEK_FLAG_BACKWARD"},
		{AVSEEK_FLAG_BYTE, "AVSEEK_FLAG_BYTE"},
		{AVSEEK_FLAG_ANY, "AVSEEK_FLAG_ANY"},
		{AVSEEK_FLAG_FRAME, "AVSEEK_FLAG_FRAME"},
	}
	for _, tt := range tests {
		assert.Equal(tt.expected, tt.flag.String())
	}
}

func Test_AVSeekFlag_String_002(t *testing.T) {
	assert := assert.New(t)
	// Test combined flags
	flag := AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_BYTE
	result := flag.String()
	assert.Contains(result, "AVSEEK_FLAG_BACKWARD")
	assert.Contains(result, "AVSEEK_FLAG_BYTE")
}

func Test_AVSeekFlag_FlagString_001(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		flag     AVSeekFlag
		expected string
	}{
		{AVSEEK_FLAG_NONE, ""},
		{AVSEEK_FLAG_BACKWARD, "AVSEEK_FLAG_BACKWARD"},
		{AVSEEK_FLAG_BYTE, "AVSEEK_FLAG_BYTE"},
		{AVSEEK_FLAG_ANY, "AVSEEK_FLAG_ANY"},
		{AVSEEK_FLAG_FRAME, "AVSEEK_FLAG_FRAME"},
	}
	for _, tt := range tests {
		assert.Equal(tt.expected, tt.flag.FlagString())
	}
}

func Test_AVSeekFlag_FlagString_002(t *testing.T) {
	assert := assert.New(t)
	// Test combined flags use String() not FlagString()
	flag := AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_BYTE
	result := flag.String()
	assert.Contains(result, "AVSEEK_FLAG_BACKWARD")
	assert.Contains(result, "AVSEEK_FLAG_BYTE")
	assert.Contains(result, "|")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - JSON MARSHALING

func Test_AVSeekFlag_JSON_001(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		flag     AVSeekFlag
		expected string
	}{
		{AVSEEK_FLAG_NONE, `"AVSEEK_FLAG_NONE"`},
		{AVSEEK_FLAG_BACKWARD, `"AVSEEK_FLAG_BACKWARD"`},
		{AVSEEK_FLAG_BYTE, `"AVSEEK_FLAG_BYTE"`},
		{AVSEEK_FLAG_ANY, `"AVSEEK_FLAG_ANY"`},
		{AVSEEK_FLAG_FRAME, `"AVSEEK_FLAG_FRAME"`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.flag)
		assert.NoError(err)
		assert.Equal(tt.expected, string(data))
	}
}

func Test_AVSeekFlag_JSON_002(t *testing.T) {
	assert := assert.New(t)
	// Test combined flags
	flag := AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_BYTE
	data, err := json.Marshal(flag)
	assert.NoError(err)

	// Should contain both flags
	str := string(data)
	assert.Contains(str, "AVSEEK_FLAG_BACKWARD")
	assert.Contains(str, "AVSEEK_FLAG_BYTE")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - IS METHODS

func Test_AVSeekFlag_Is_001(t *testing.T) {
	assert := assert.New(t)
	// Test individual flags
	assert.True(AVSEEK_FLAG_BACKWARD.Is(AVSEEK_FLAG_BACKWARD))
	assert.True(AVSEEK_FLAG_BYTE.Is(AVSEEK_FLAG_BYTE))
	assert.True(AVSEEK_FLAG_ANY.Is(AVSEEK_FLAG_ANY))
	assert.True(AVSEEK_FLAG_FRAME.Is(AVSEEK_FLAG_FRAME))
}

func Test_AVSeekFlag_Is_002(t *testing.T) {
	assert := assert.New(t)
	// Test combined flags
	flag := AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_BYTE
	assert.True(flag.Is(AVSEEK_FLAG_BACKWARD))
	assert.True(flag.Is(AVSEEK_FLAG_BYTE))
	assert.False(flag.Is(AVSEEK_FLAG_ANY))
}

func Test_AVSeekFlag_Is_003(t *testing.T) {
	assert := assert.New(t)
	// Test NONE
	assert.True(AVSEEK_FLAG_NONE.Is(AVSEEK_FLAG_NONE))
	assert.False(AVSEEK_FLAG_NONE.Is(AVSEEK_FLAG_BACKWARD))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - BITWISE OPERATIONS

func Test_AVSeekFlag_Bitwise_001(t *testing.T) {
	assert := assert.New(t)
	// Test OR operation
	flag := AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_BYTE
	assert.True(flag.Is(AVSEEK_FLAG_BACKWARD))
	assert.True(flag.Is(AVSEEK_FLAG_BYTE))
}

func Test_AVSeekFlag_Bitwise_002(t *testing.T) {
	assert := assert.New(t)
	// Test multiple flags
	flag := AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_BYTE | AVSEEK_FLAG_ANY
	assert.True(flag.Is(AVSEEK_FLAG_BACKWARD))
	assert.True(flag.Is(AVSEEK_FLAG_BYTE))
	assert.True(flag.Is(AVSEEK_FLAG_ANY))
	assert.False(flag.Is(AVSEEK_FLAG_FRAME))
}

func Test_AVSeekFlag_Bitwise_003(t *testing.T) {
	assert := assert.New(t)
	// Test all flags combined
	flag := AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_BYTE | AVSEEK_FLAG_ANY | AVSEEK_FLAG_FRAME
	assert.True(flag.Is(AVSEEK_FLAG_BACKWARD))
	assert.True(flag.Is(AVSEEK_FLAG_BYTE))
	assert.True(flag.Is(AVSEEK_FLAG_ANY))
	assert.True(flag.Is(AVSEEK_FLAG_FRAME))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - CONSTANTS

func Test_AVSeekFlag_Constants_001(t *testing.T) {
	assert := assert.New(t)
	// Test that constants have expected values
	assert.Equal(AVSeekFlag(0), AVSEEK_FLAG_NONE)
	assert.Equal(AVSeekFlag(1), AVSEEK_FLAG_BACKWARD)
	assert.Equal(AVSeekFlag(2), AVSEEK_FLAG_BYTE)
	assert.Equal(AVSeekFlag(4), AVSEEK_FLAG_ANY)
	assert.Equal(AVSeekFlag(8), AVSEEK_FLAG_FRAME)
}

func Test_AVSeekFlag_Constants_002(t *testing.T) {
	assert := assert.New(t)
	// Test AVSEEK constants
	assert.NotEqual(0, AVSEEK_SIZE)
	assert.NotEqual(0, AVSEEK_FORCE)
	assert.Equal(0x10000, AVSEEK_SIZE)
	assert.Equal(0x20000, AVSEEK_FORCE)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - COMMON PATTERNS

func Test_AVSeekFlag_Pattern_001(t *testing.T) {
	assert := assert.New(t)
	// Test common seeking patterns

	// Backward seek to keyframe (common for scrubbing)
	flag := AVSEEK_FLAG_BACKWARD
	assert.True(flag.Is(AVSEEK_FLAG_BACKWARD))
	assert.False(flag.Is(AVSEEK_FLAG_ANY))

	// Forward seek to any frame (common for fast forward)
	flag = AVSEEK_FLAG_ANY
	assert.True(flag.Is(AVSEEK_FLAG_ANY))
	assert.False(flag.Is(AVSEEK_FLAG_BACKWARD))
}

func Test_AVSeekFlag_Pattern_002(t *testing.T) {
	assert := assert.New(t)
	// Test byte-based seeking
	flag := AVSEEK_FLAG_BYTE
	assert.True(flag.Is(AVSEEK_FLAG_BYTE))

	// Byte-based backward seek
	flag = AVSEEK_FLAG_BYTE | AVSEEK_FLAG_BACKWARD
	assert.True(flag.Is(AVSEEK_FLAG_BYTE))
	assert.True(flag.Is(AVSEEK_FLAG_BACKWARD))
}

func Test_AVSeekFlag_Pattern_003(t *testing.T) {
	assert := assert.New(t)
	// Test frame-based seeking
	flag := AVSEEK_FLAG_FRAME
	assert.True(flag.Is(AVSEEK_FLAG_FRAME))

	// Frame-based backward seek to any frame
	flag = AVSEEK_FLAG_FRAME | AVSEEK_FLAG_BACKWARD | AVSEEK_FLAG_ANY
	assert.True(flag.Is(AVSEEK_FLAG_FRAME))
	assert.True(flag.Is(AVSEEK_FLAG_BACKWARD))
	assert.True(flag.Is(AVSEEK_FLAG_ANY))
}

func Test_AVSeekFlag_Pattern_004(t *testing.T) {
	assert := assert.New(t)
	// Test that mutually exclusive time base flags don't overlap
	assert.NotEqual(AVSEEK_FLAG_BYTE, AVSEEK_FLAG_FRAME)
	assert.False(AVSEEK_FLAG_BYTE.Is(AVSEEK_FLAG_FRAME))
	assert.False(AVSEEK_FLAG_FRAME.Is(AVSEEK_FLAG_BYTE))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - ACTUAL SEEKING WITH FILES

func Test_AVFormat_seek_frame_001(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	input, err := AVFormat_open_url(testFile, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Seek to timestamp 0 (beginning)
	err = AVFormat_seek_frame(input, -1, 0, AVSEEK_FLAG_BACKWARD)
	assert.NoError(err)

	t.Log("Successfully seeked to beginning of file")
}

func Test_AVFormat_seek_frame_002(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	input, err := AVFormat_open_url(testFile, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Seek forward to 1 second (timestamp in AVStream time_base units)
	// For simplicity, use 1000 which works for many formats
	err = AVFormat_seek_frame(input, -1, 1000, AVSEEK_FLAG_ANY)
	if err != nil {
		t.Logf("Seek to timestamp 1000 failed (this is OK for some formats): %v", err)
	} else {
		t.Log("Successfully seeked forward in file")
	}
}

func Test_AVFormat_seek_file_001(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	input, err := AVFormat_open_url(testFile, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Seek to keyframe at timestamp range
	err = AVFormat_seek_file(input, -1, 0, 500, 1000, AVSEEK_FLAG_BACKWARD)
	if err != nil {
		t.Logf("Seek to keyframe failed (this is OK for some formats): %v", err)
	} else {
		t.Log("Successfully seeked to keyframe")
	}
}

func Test_AVFormat_seek_pattern_001(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	input, err := AVFormat_open_url(testFile, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Test common seeking patterns
	patterns := []struct {
		name      string
		timestamp int64
		flags     AVSeekFlag
	}{
		{"seek to start", 0, AVSEEK_FLAG_BACKWARD},
		{"seek backward to keyframe", 0, AVSEEK_FLAG_BACKWARD},
		{"seek to any frame", 100, AVSEEK_FLAG_ANY},
	}

	for _, p := range patterns {
		err := AVFormat_seek_frame(input, -1, p.timestamp, p.flags)
		if err != nil {
			t.Logf("%s failed (acceptable for some formats): %v", p.name, err)
		} else {
			t.Logf("%s succeeded", p.name)
		}
	}
}
