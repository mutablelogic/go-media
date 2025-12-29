package ffmpeg

import (
	"encoding/json"
	"strings"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST STRING OUTPUT

func Test_avformat_disposition_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		disposition AVDisposition
		expected    string
	}{
		{AV_DISPOSITION_DEFAULT, "DEFAULT"},
		{AV_DISPOSITION_DUB, "DUB"},
		{AV_DISPOSITION_ORIGINAL, "ORIGINAL"},
		{AV_DISPOSITION_COMMENT, "COMMENT"},
		{AV_DISPOSITION_LYRICS, "LYRICS"},
		{AV_DISPOSITION_KARAOKE, "KARAOKE"},
		{AV_DISPOSITION_FORCED, "FORCED"},
		{AV_DISPOSITION_HEARING_IMPAIRED, "HEARING_IMPAIRED"},
		{AV_DISPOSITION_VISUAL_IMPAIRED, "VISUAL_IMPAIRED"},
		{AV_DISPOSITION_CLEAN_EFFECTS, "CLEAN_EFFECTS"},
		{AV_DISPOSITION_ATTACHED_PIC, "ATTACHED_PIC"},
		{AV_DISPOSITION_TIMED_THUMBNAILS, "TIMED_THUMBNAILS"},
		{AV_DISPOSITION_NON_DIEGETIC, "NON_DIEGETIC"},
		{AV_DISPOSITION_CAPTIONS, "CAPTIONS"},
		{AV_DISPOSITION_DESCRIPTIONS, "DESCRIPTIONS"},
		{AV_DISPOSITION_METADATA, "METADATA"},
		{AV_DISPOSITION_DEPENDENT, "DEPENDENT"},
		{AV_DISPOSITION_STILL_IMAGE, "STILL_IMAGE"},
		{AV_DISPOSITION_MULTILAYER, "MULTILAYER"},
	}

	for _, tc := range tests {
		str := tc.disposition.String()
		assert.Equal(tc.expected, str)
		t.Logf("Disposition %v: %q", tc.disposition, str)
	}
}

func Test_avformat_disposition_string_zero(t *testing.T) {
	assert := assert.New(t)

	// Zero disposition should return empty string
	var zero AVDisposition = 0
	str := zero.String()
	assert.Equal("", str)
	t.Logf("Zero disposition: %q", str)
}

func Test_avformat_disposition_string_combined(t *testing.T) {
	assert := assert.New(t)

	// Test combined flags
	combined := AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED
	str := combined.String()
	assert.Contains(str, "DEFAULT")
	assert.Contains(str, "FORCED")
	assert.Contains(str, "|")
	t.Logf("Combined disposition (DEFAULT|FORCED): %q", str)

	// Test multiple flags
	multi := AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED | AV_DISPOSITION_CAPTIONS
	str = multi.String()
	parts := strings.Split(str, "|")
	assert.Len(parts, 3)
	assert.Contains(str, "DEFAULT")
	assert.Contains(str, "FORCED")
	assert.Contains(str, "CAPTIONS")
	t.Logf("Multi disposition (DEFAULT|FORCED|CAPTIONS): %q", str)
}

func Test_avformat_disposition_flagstring(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		disposition AVDisposition
		expected    string
	}{
		{AV_DISPOSITION_DEFAULT, "DEFAULT"},
		{AV_DISPOSITION_DUB, "DUB"},
		{AV_DISPOSITION_ORIGINAL, "ORIGINAL"},
		{AV_DISPOSITION_FORCED, "FORCED"},
		{AV_DISPOSITION_CAPTIONS, "CAPTIONS"},
	}

	for _, tc := range tests {
		str := tc.disposition.FlagString()
		assert.Equal(tc.expected, str)
		t.Logf("FlagString for %v: %q", tc.disposition, str)
	}
}

func Test_avformat_disposition_flagstring_invalid(t *testing.T) {
	assert := assert.New(t)

	// Test with invalid disposition value (not a known flag)
	invalidDisp := AVDisposition(0x12345678)
	str := invalidDisp.FlagString()
	assert.NotEmpty(str)
	assert.Contains(str, "AVDisposition")
	assert.Contains(str, "0x12345678")
	t.Logf("Invalid disposition FlagString: %q", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avformat_disposition_json_marshaling(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		disposition AVDisposition
		expected    string
	}{
		{AV_DISPOSITION_DEFAULT, `"DEFAULT"`},
		{AV_DISPOSITION_DUB, `"DUB"`},
		{AV_DISPOSITION_FORCED, `"FORCED"`},
		{AV_DISPOSITION_CAPTIONS, `"CAPTIONS"`},
		{0, `""`}, // Zero should marshal to empty string
	}

	for _, tc := range tests {
		jsonBytes, err := json.Marshal(tc.disposition)
		assert.NoError(err)
		assert.Equal(tc.expected, string(jsonBytes))
		t.Logf("Disposition %v marshals to: %s", tc.disposition, string(jsonBytes))
	}
}

func Test_avformat_disposition_json_combined(t *testing.T) {
	assert := assert.New(t)

	// Test combined flags in JSON
	combined := AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED
	jsonBytes, err := json.Marshal(combined)
	assert.NoError(err)

	var jsonStr string
	err = json.Unmarshal(jsonBytes, &jsonStr)
	assert.NoError(err)
	assert.Contains(jsonStr, "DEFAULT")
	assert.Contains(jsonStr, "FORCED")
	assert.Contains(jsonStr, "|")
	t.Logf("Combined disposition JSON: %s", string(jsonBytes))
}

func Test_avformat_disposition_json_in_struct(t *testing.T) {
	assert := assert.New(t)

	type StreamInfo struct {
		Disposition AVDisposition `json:"disposition"`
		Index       int           `json:"index"`
	}

	tests := []struct {
		name     string
		input    StreamInfo
		expected string
	}{
		{
			name:     "default",
			input:    StreamInfo{Disposition: AV_DISPOSITION_DEFAULT, Index: 0},
			expected: `{"disposition":"DEFAULT","index":0}`,
		},
		{
			name:     "forced",
			input:    StreamInfo{Disposition: AV_DISPOSITION_FORCED, Index: 1},
			expected: `{"disposition":"FORCED","index":1}`,
		},
		{
			name:     "combined",
			input:    StreamInfo{Disposition: AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED, Index: 2},
			expected: `DEFAULT|FORCED`, // Check it contains both
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tc.input)
			assert.NoError(err)
			jsonStr := string(jsonBytes)

			if strings.Contains(tc.expected, "|") {
				// For combined flags, just check both are present
				assert.Contains(jsonStr, "DEFAULT")
				assert.Contains(jsonStr, "FORCED")
			} else {
				assert.Equal(tc.expected, jsonStr)
			}
			t.Logf("%s JSON: %s", tc.name, jsonStr)
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST IS METHOD

func Test_avformat_disposition_is_method(t *testing.T) {
	assert := assert.New(t)

	// Test single flag checking
	assert.True(AV_DISPOSITION_DEFAULT.Is(AV_DISPOSITION_DEFAULT))
	assert.False(AV_DISPOSITION_DEFAULT.Is(AV_DISPOSITION_FORCED))
	assert.False(AV_DISPOSITION_FORCED.Is(AV_DISPOSITION_DEFAULT))

	// Test combined flags
	combined := AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED
	assert.True(combined.Is(AV_DISPOSITION_DEFAULT))
	assert.True(combined.Is(AV_DISPOSITION_FORCED))
	assert.False(combined.Is(AV_DISPOSITION_CAPTIONS))

	// Test checking multiple flags
	assert.True(combined.Is(AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED))
	assert.False(combined.Is(AV_DISPOSITION_DEFAULT | AV_DISPOSITION_CAPTIONS))

	t.Log("Is() method tests passed")
}

func Test_avformat_disposition_is_with_multiple_flags(t *testing.T) {
	assert := assert.New(t)

	// Test with multiple flags set
	multi := AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED | AV_DISPOSITION_CAPTIONS

	// Individual flag checks
	assert.True(multi.Is(AV_DISPOSITION_DEFAULT))
	assert.True(multi.Is(AV_DISPOSITION_FORCED))
	assert.True(multi.Is(AV_DISPOSITION_CAPTIONS))
	assert.False(multi.Is(AV_DISPOSITION_DUB))

	// Multi-flag checks
	assert.True(multi.Is(AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED))
	assert.True(multi.Is(AV_DISPOSITION_FORCED | AV_DISPOSITION_CAPTIONS))
	assert.True(multi.Is(AV_DISPOSITION_DEFAULT | AV_DISPOSITION_CAPTIONS))
	assert.True(multi.Is(AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED | AV_DISPOSITION_CAPTIONS))

	// Check with flag not present
	assert.False(multi.Is(AV_DISPOSITION_DUB))
	assert.False(multi.Is(AV_DISPOSITION_DEFAULT | AV_DISPOSITION_DUB))

	t.Log("Multiple flags Is() tests passed")
}

////////////////////////////////////////////////////////////////////////////////
// TEST CONSTANTS VALUES

func Test_avformat_disposition_constants_values(t *testing.T) {
	assert := assert.New(t)

	// Verify constants are powers of 2 (single bit set)
	dispositions := []AVDisposition{
		AV_DISPOSITION_DEFAULT,
		AV_DISPOSITION_DUB,
		AV_DISPOSITION_ORIGINAL,
		AV_DISPOSITION_COMMENT,
		AV_DISPOSITION_LYRICS,
		AV_DISPOSITION_KARAOKE,
		AV_DISPOSITION_FORCED,
		AV_DISPOSITION_HEARING_IMPAIRED,
		AV_DISPOSITION_VISUAL_IMPAIRED,
		AV_DISPOSITION_CLEAN_EFFECTS,
		AV_DISPOSITION_ATTACHED_PIC,
		AV_DISPOSITION_TIMED_THUMBNAILS,
		AV_DISPOSITION_NON_DIEGETIC,
		AV_DISPOSITION_CAPTIONS,
		AV_DISPOSITION_DESCRIPTIONS,
		AV_DISPOSITION_METADATA,
		AV_DISPOSITION_DEPENDENT,
		AV_DISPOSITION_STILL_IMAGE,
		AV_DISPOSITION_MULTILAYER,
	}

	for _, disp := range dispositions {
		// Check that it's a power of 2 (has only one bit set)
		// A power of 2 & (power of 2 - 1) should be 0
		assert.Equal(AVDisposition(0), disp&(disp-1), "Disposition %v should be a power of 2", disp)
		t.Logf("Disposition %s = 0x%X (power of 2)", disp.String(), int(disp))
	}
}

func Test_avformat_disposition_constants_unique(t *testing.T) {
	assert := assert.New(t)

	dispositions := []AVDisposition{
		AV_DISPOSITION_DEFAULT,
		AV_DISPOSITION_DUB,
		AV_DISPOSITION_ORIGINAL,
		AV_DISPOSITION_COMMENT,
		AV_DISPOSITION_LYRICS,
		AV_DISPOSITION_KARAOKE,
		AV_DISPOSITION_FORCED,
		AV_DISPOSITION_HEARING_IMPAIRED,
		AV_DISPOSITION_VISUAL_IMPAIRED,
		AV_DISPOSITION_CLEAN_EFFECTS,
		AV_DISPOSITION_ATTACHED_PIC,
		AV_DISPOSITION_TIMED_THUMBNAILS,
		AV_DISPOSITION_NON_DIEGETIC,
		AV_DISPOSITION_CAPTIONS,
		AV_DISPOSITION_DESCRIPTIONS,
		AV_DISPOSITION_METADATA,
		AV_DISPOSITION_DEPENDENT,
		AV_DISPOSITION_STILL_IMAGE,
		AV_DISPOSITION_MULTILAYER,
	}

	// Check all values are unique
	seen := make(map[AVDisposition]bool)
	for _, disp := range dispositions {
		assert.False(seen[disp], "Disposition %v appears multiple times", disp)
		seen[disp] = true
	}

	assert.Equal(len(dispositions), len(seen))
	t.Logf("All %d disposition constants are unique", len(seen))
}

////////////////////////////////////////////////////////////////////////////////
// TEST BITWISE OPERATIONS

func Test_avformat_disposition_bitwise_operations(t *testing.T) {
	assert := assert.New(t)

	// Test OR operation (combining flags)
	combined := AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED
	assert.True(combined.Is(AV_DISPOSITION_DEFAULT))
	assert.True(combined.Is(AV_DISPOSITION_FORCED))
	t.Logf("OR operation: %s", combined.String())

	// Test AND operation (checking flags)
	result := combined & AV_DISPOSITION_DEFAULT
	assert.Equal(AV_DISPOSITION_DEFAULT, result)
	t.Logf("AND with DEFAULT: %s", result.String())

	// Test XOR operation (toggling flags)
	toggled := combined ^ AV_DISPOSITION_DEFAULT
	assert.False(toggled.Is(AV_DISPOSITION_DEFAULT))
	assert.True(toggled.Is(AV_DISPOSITION_FORCED))
	t.Logf("XOR toggle DEFAULT: %s", toggled.String())

	// Test NOT operation with AND (clearing flags)
	cleared := combined &^ AV_DISPOSITION_DEFAULT
	assert.False(cleared.Is(AV_DISPOSITION_DEFAULT))
	assert.True(cleared.Is(AV_DISPOSITION_FORCED))
	t.Logf("Clear DEFAULT: %s", cleared.String())
}

////////////////////////////////////////////////////////////////////////////////
// TEST COMMON USE CASES

func Test_avformat_disposition_common_patterns(t *testing.T) {
	assert := assert.New(t)

	// Pattern 1: Check if stream is default
	var streamDisp AVDisposition = AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED
	if streamDisp.Is(AV_DISPOSITION_DEFAULT) {
		t.Log("Stream is marked as default")
	}
	assert.True(streamDisp.Is(AV_DISPOSITION_DEFAULT))

	// Pattern 2: Check if subtitle is forced
	var subtitleDisp AVDisposition = AV_DISPOSITION_FORCED
	if subtitleDisp.Is(AV_DISPOSITION_FORCED) {
		t.Log("Subtitle is forced")
	}
	assert.True(subtitleDisp.Is(AV_DISPOSITION_FORCED))

	// Pattern 3: Check for accessibility features
	var audioDisp AVDisposition = AV_DISPOSITION_HEARING_IMPAIRED | AV_DISPOSITION_DESCRIPTIONS
	hasAccessibility := audioDisp.Is(AV_DISPOSITION_HEARING_IMPAIRED) ||
		audioDisp.Is(AV_DISPOSITION_VISUAL_IMPAIRED) ||
		audioDisp.Is(AV_DISPOSITION_DESCRIPTIONS)
	assert.True(hasAccessibility)
	t.Log("Stream has accessibility features")

	// Pattern 4: Check if stream is a cover art
	var imageDisp AVDisposition = AV_DISPOSITION_ATTACHED_PIC
	if imageDisp.Is(AV_DISPOSITION_ATTACHED_PIC) {
		t.Log("Stream is attached picture (cover art)")
	}
	assert.True(imageDisp.Is(AV_DISPOSITION_ATTACHED_PIC))
}

func Test_avformat_disposition_filter_streams(t *testing.T) {
	assert := assert.New(t)

	// Simulate filtering streams by disposition
	type Stream struct {
		Index       int
		Disposition AVDisposition
	}

	streams := []Stream{
		{0, AV_DISPOSITION_DEFAULT},
		{1, AV_DISPOSITION_FORCED},
		{2, AV_DISPOSITION_DEFAULT | AV_DISPOSITION_FORCED},
		{3, AV_DISPOSITION_CAPTIONS},
		{4, 0}, // No disposition
	}

	// Find default streams
	var defaultStreams []Stream
	for _, s := range streams {
		if s.Disposition.Is(AV_DISPOSITION_DEFAULT) {
			defaultStreams = append(defaultStreams, s)
		}
	}
	assert.Len(defaultStreams, 2)
	t.Logf("Found %d default streams", len(defaultStreams))

	// Find forced streams
	var forcedStreams []Stream
	for _, s := range streams {
		if s.Disposition.Is(AV_DISPOSITION_FORCED) {
			forcedStreams = append(forcedStreams, s)
		}
	}
	assert.Len(forcedStreams, 2)
	t.Logf("Found %d forced streams", len(forcedStreams))

	// Find streams with no disposition
	var noDispositionStreams []Stream
	for _, s := range streams {
		if s.Disposition == 0 {
			noDispositionStreams = append(noDispositionStreams, s)
		}
	}
	assert.Len(noDispositionStreams, 1)
	t.Logf("Found %d streams with no disposition", len(noDispositionStreams))
}
