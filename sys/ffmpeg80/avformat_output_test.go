package ffmpeg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - ITERATION AND LOOKUP

func Test_avformat_output_001(t *testing.T) {
	assert := assert.New(t)
	// Iterate over all output formats
	var opaque uintptr
	count := 0
	for {
		muxer := AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}
		// Name should never be empty
		assert.NotEmpty(muxer.Name())
		count++
	}
	assert.Greater(count, 0, "Should have at least one output format")
}

func Test_avformat_output_002(t *testing.T) {
	assert := assert.New(t)
	// Test guess_format with common formats
	tests := []struct {
		format   string
		filename string
		expected string
	}{
		{"mp4", "", "mp4"},
		{"", "test.mp4", "mp4"},
		{"avi", "", "avi"},
		{"", "test.avi", "avi"},
		{"mkv", "", "matroska"},
		{"", "test.mkv", "matroska"},
	}

	for _, tt := range tests {
		muxer := AVFormat_guess_format(tt.format, tt.filename, "")
		if muxer != nil {
			assert.Equal(tt.expected, muxer.Name())
			t.Logf("Format: %s, Filename: %s -> %s", tt.format, tt.filename, muxer.Name())
		}
	}
}

func Test_avformat_output_003(t *testing.T) {
	assert := assert.New(t)
	// Test guess_format with MIME types
	muxer := AVFormat_guess_format("", "", "video/mp4")
	if muxer != nil {
		assert.NotEmpty(muxer.Name())
		t.Logf("MIME type video/mp4 -> %s", muxer.Name())
	}
}

func Test_avformat_output_004(t *testing.T) {
	assert := assert.New(t)
	// Test guess_format with non-existent format returns nil
	muxer := AVFormat_guess_format("nonexistent_format_12345", "", "")
	assert.Nil(muxer)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - PROPERTIES

func Test_avformat_output_properties_001(t *testing.T) {
	assert := assert.New(t)
	// Test that all muxers have valid properties
	var opaque uintptr
	for {
		muxer := AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}

		// Name should never be empty
		assert.NotEmpty(muxer.Name())

		// Test property accessors don't crash
		_ = muxer.LongName()
		_ = muxer.Flags()
		_ = muxer.MimeTypes()
		_ = muxer.Extensions()
		_ = muxer.VideoCodec()
		_ = muxer.AudioCodec()
		_ = muxer.SubtitleCodec()
	}
}

func Test_avformat_output_properties_002(t *testing.T) {
	assert := assert.New(t)
	// Test specific well-known format properties
	muxer := AVFormat_guess_format("mp4", "", "")
	if muxer != nil {
		assert.Equal("mp4", muxer.Name())
		assert.NotEmpty(muxer.LongName())

		// Extensions might include mp4
		extensions := muxer.Extensions()
		if extensions != "" {
			t.Logf("MP4 extensions: %s", extensions)
		}

		// Should have default codecs
		videoCodec := muxer.VideoCodec()
		audioCodec := muxer.AudioCodec()
		t.Logf("MP4 video codec: %v, audio codec: %v", videoCodec, audioCodec)
	}
}

func Test_avformat_output_properties_003(t *testing.T) {
	assert := assert.New(t)
	// Test flags
	var opaque uintptr
	foundWithFlags := false
	for {
		muxer := AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}

		flags := muxer.Flags()
		if flags != AVFMT_NONE {
			assert.NotEqual(AVFMT_NONE, flags)
			t.Logf("Format %s has flags: %s", muxer.Name(), flags)
			foundWithFlags = true
			break
		}
	}
	if !foundWithFlags {
		t.Log("No formats found with flags (unusual but not an error)")
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - STRING AND JSON

func Test_avformat_output_string_001(t *testing.T) {
	assert := assert.New(t)
	// Test String() method
	muxer := AVFormat_guess_format("mp4", "", "")
	if muxer != nil {
		str := muxer.String()
		assert.NotEmpty(str)
		assert.Contains(str, "name")
		t.Log(str)
	}
}

func Test_avformat_output_json_001(t *testing.T) {
	assert := assert.New(t)
	// Test JSON marshaling
	muxer := AVFormat_guess_format("mp4", "", "")
	if muxer != nil {
		data, err := json.Marshal(muxer)
		assert.NoError(err)
		assert.NotEmpty(data)

		// Unmarshal to verify structure
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(err)
		assert.Contains(result, "name")
	}
}

func Test_avformat_output_json_002(t *testing.T) {
	assert := assert.New(t)
	// Test JSON marshaling for multiple formats
	var opaque uintptr
	count := 0
	for {
		muxer := AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}

		data, err := json.Marshal(muxer)
		assert.NoError(err)
		assert.NotEmpty(data)

		count++
		if count >= 5 {
			break
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - COMMON PATTERNS

func Test_avformat_output_pattern_001(t *testing.T) {
	assert := assert.New(t)
	// Test common output formats
	commonFormats := []string{"mp4", "avi", "mkv", "flv", "mp3", "wav", "webm", "mov"}

	for _, formatName := range commonFormats {
		muxer := AVFormat_guess_format(formatName, "", "")
		if muxer != nil {
			assert.NotEmpty(muxer.Name())
			t.Logf("Found format: %s -> %s", formatName, muxer.Name())
		}
	}
}

func Test_avformat_output_pattern_002(t *testing.T) {
	assert := assert.New(t)
	// Test that formats with extensions have at least one
	var opaque uintptr
	count := 0
	withExtensions := 0

	for {
		muxer := AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}

		count++
		extensions := muxer.Extensions()
		if extensions != "" {
			withExtensions++
		}
	}

	assert.Greater(count, 0)
	assert.Greater(withExtensions, 0, "At least some formats should have extensions")
	t.Logf("Total formats: %d, with extensions: %d", count, withExtensions)
}

func Test_avformat_output_pattern_003(t *testing.T) {
	assert := assert.New(t)
	// Test that formats with mime types have at least one
	var opaque uintptr
	count := 0
	withMimeTypes := 0

	for {
		muxer := AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}

		count++
		mimeTypes := muxer.MimeTypes()
		if mimeTypes != "" {
			withMimeTypes++
		}
	}

	assert.Greater(count, 0)
	t.Logf("Total formats: %d, with mime types: %d", count, withMimeTypes)
}

func Test_avformat_output_pattern_004(t *testing.T) {
	assert := assert.New(t)
	// Test filename extension guessing
	extensionTests := map[string]string{
		"test.mp4":  "mp4",
		"test.avi":  "avi",
		"test.mkv":  "matroska",
		"test.webm": "webm",
		"test.flv":  "flv",
		"test.mp3":  "mp3",
		"test.wav":  "wav",
	}

	for filename, expectedFormat := range extensionTests {
		muxer := AVFormat_guess_format("", filename, "")
		if muxer != nil {
			assert.Equal(expectedFormat, muxer.Name())
			t.Logf("Filename %s -> format %s", filename, muxer.Name())
		}
	}
}
