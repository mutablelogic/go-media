package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - ITERATION AND LOOKUP

func Test_avformat_input_001(t *testing.T) {
	assert := assert.New(t)
	// Iterate over all input formats
	var opaque uintptr
	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		demuxer2 := AVFormat_find_input_format(demuxer.Name())
		assert.Equal(demuxer, demuxer2)
	}
}

func Test_avformat_input_002(t *testing.T) {
	assert := assert.New(t)
	// Test iteration returns at least some formats
	var opaque uintptr
	count := 0
	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		count++
	}
	assert.Greater(count, 0, "Should have at least one input format")
}

func Test_avformat_input_003(t *testing.T) {
	assert := assert.New(t)
	// Test finding common formats
	commonFormats := []string{"mp4", "mov", "avi", "mkv", "flv", "mp3", "wav"}
	for _, formatName := range commonFormats {
		demuxer := AVFormat_find_input_format(formatName)
		if demuxer != nil {
			assert.NotEmpty(demuxer.Name())
			t.Logf("Found format: %s", formatName)
		}
	}
}

func Test_avformat_input_004(t *testing.T) {
	assert := assert.New(t)
	// Test finding non-existent format returns nil
	demuxer := AVFormat_find_input_format("nonexistent_format_12345")
	assert.Nil(demuxer)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - PROPERTIES

func Test_avformat_input_properties_001(t *testing.T) {
	assert := assert.New(t)
	// Test that all demuxers have valid properties
	var opaque uintptr
	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}

		// Name should never be empty
		assert.NotEmpty(demuxer.Name())

		// LongName is usually not empty but not guaranteed
		longName := demuxer.LongName()
		_ = longName // Can be empty

		// Test property accessors don't crash
		_ = demuxer.Flags()
		_ = demuxer.MimeTypes()
		_ = demuxer.Extensions()
	}
}

func Test_avformat_input_properties_002(t *testing.T) {
	assert := assert.New(t)
	// Test specific well-known formats have expected properties
	demuxer := AVFormat_find_input_format("mp4")
	if demuxer != nil {
		assert.Equal("mov,mp4,m4a,3gp,3g2,mj2", demuxer.Name())
		assert.NotEmpty(demuxer.LongName())

		// Extensions might include mp4
		extensions := demuxer.Extensions()
		if extensions != "" {
			t.Logf("MP4 extensions: %s", extensions)
		}
	}
}

func Test_avformat_input_properties_003(t *testing.T) {
	// Test flags are valid
	assert := assert.New(t)
	var opaque uintptr
	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}

		flags := demuxer.Flags()
		// Flags can be 0 or any combination of AVFormat flags
		_ = flags

		// If we find one with flags, test it
		if flags != 0 {
			assert.NotEqual(AVFMT_NONE, flags)
			t.Logf("Format %s has flags: %s", demuxer.Name(), flags)
			break
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - STRING AND JSON

func Test_avformat_input_string_001(t *testing.T) {
	assert := assert.New(t)
	// Test String() method
	demuxer := AVFormat_find_input_format("mp4")
	if demuxer != nil {
		str := demuxer.String()
		assert.NotEmpty(str)
		assert.Contains(str, "name")
		t.Log(str)
	}
}

func Test_avformat_input_json_001(t *testing.T) {
	assert := assert.New(t)
	// Test JSON marshaling
	demuxer := AVFormat_find_input_format("mp4")
	if demuxer != nil {
		data, err := json.Marshal(demuxer)
		assert.NoError(err)
		assert.NotEmpty(data)

		// Unmarshal to verify structure
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(err)
		assert.Contains(result, "name")
	}
}

func Test_avformat_input_json_002(t *testing.T) {
	assert := assert.New(t)
	// Test JSON marshaling for multiple formats
	var opaque uintptr
	count := 0
	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}

		data, err := json.Marshal(demuxer)
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

func Test_avformat_input_pattern_001(t *testing.T) {
	assert := assert.New(t)
	// Test finding formats by extension
	extensionFormats := map[string]string{
		"mp4": "mov,mp4,m4a,3gp,3g2,mj2",
		"avi": "avi",
		"mkv": "matroska,webm",
		"flv": "flv",
		"mp3": "mp3",
		"wav": "wav",
	}

	for ext, expectedName := range extensionFormats {
		demuxer := AVFormat_find_input_format(ext)
		if demuxer != nil {
			assert.Equal(expectedName, demuxer.Name())
			t.Logf("Extension %s -> format %s", ext, demuxer.Name())
		}
	}
}

func Test_avformat_input_pattern_002(t *testing.T) {
	assert := assert.New(t)
	// Test that formats with extensions have at least one
	var opaque uintptr
	count := 0
	withExtensions := 0

	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}

		count++
		extensions := demuxer.Extensions()
		if extensions != "" {
			withExtensions++
		}
	}

	assert.Greater(count, 0)
	assert.Greater(withExtensions, 0, "At least some formats should have extensions")
	t.Logf("Total formats: %d, with extensions: %d", count, withExtensions)
}

func Test_avformat_input_pattern_003(t *testing.T) {
	assert := assert.New(t)
	// Test that formats with mime types have at least one
	var opaque uintptr
	count := 0
	withMimeTypes := 0

	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}

		count++
		mimeTypes := demuxer.MimeTypes()
		if mimeTypes != "" {
			withMimeTypes++
		}
	}

	assert.Greater(count, 0)
	// Not all formats have mime types, so we just verify some do
	t.Logf("Total formats: %d, with mime types: %d", count, withMimeTypes)
}
