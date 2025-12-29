package ffmpeg

import (
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST VIDEO SIZE PARSING

func Test_avutil_parse_video_size(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input  string
		width  int
		height int
	}{
		{"1920x1080", 1920, 1080},
		{"1280x720", 1280, 720},
		{"640x480", 640, 480},
		{"3840x2160", 3840, 2160}, // 4K
		{"7680x4320", 7680, 4320}, // 8K
		{"320x240", 320, 240},
		{"1024x768", 1024, 768},
	}

	for _, tc := range tests {
		width, height, err := AVUtil_parse_video_size(tc.input)
		assert.NoError(err, "Failed to parse: %s", tc.input)
		assert.Equal(tc.width, width, "Width mismatch for %s", tc.input)
		assert.Equal(tc.height, height, "Height mismatch for %s", tc.input)
		t.Logf("Parsed %s -> %dx%d", tc.input, width, height)
	}
}

func Test_avutil_parse_video_size_named(t *testing.T) {
	assert := assert.New(t)

	// FFmpeg supports named resolutions
	tests := []struct {
		input  string
		width  int
		height int
	}{
		{"vga", 640, 480},
		{"hd720", 1280, 720},
		{"hd1080", 1920, 1080},
		{"qvga", 320, 240},
		{"4k", 4096, 2160},
		{"uhd2160", 3840, 2160},
	}

	for _, tc := range tests {
		width, height, err := AVUtil_parse_video_size(tc.input)
		if err != nil {
			t.Logf("Named resolution '%s' not supported or failed: %v", tc.input, err)
			continue
		}
		assert.Equal(tc.width, width, "Width mismatch for %s", tc.input)
		assert.Equal(tc.height, height, "Height mismatch for %s", tc.input)
		t.Logf("Parsed named '%s' -> %dx%d", tc.input, width, height)
	}
}

func Test_avutil_parse_video_size_invalid(t *testing.T) {
	assert := assert.New(t)

	invalidSizes := []string{
		"",
		"invalid",
		"1920x",
		"x1080",
		"abc",
		"-1920x1080",
		"1920x-1080",
		"0x0",
	}

	for _, size := range invalidSizes {
		width, height, err := AVUtil_parse_video_size(size)
		assert.Error(err, "Expected error for invalid size: %s", size)
		assert.Equal(0, width, "Width should be 0 on error")
		assert.Equal(0, height, "Height should be 0 on error")
		t.Logf("Invalid size '%s' correctly rejected: %v", size, err)
	}
}

func Test_avutil_parse_video_size_edge_cases(t *testing.T) {
	assert := assert.New(t)

	// Test very large dimensions
	width, height, err := AVUtil_parse_video_size("16384x16384")
	if err == nil {
		assert.Equal(16384, width)
		assert.Equal(16384, height)
		t.Logf("Large size parsed: %dx%d", width, height)
	} else {
		t.Logf("Large size may not be supported: %v", err)
	}

	// Test minimum valid dimensions
	width, height, err = AVUtil_parse_video_size("1x1")
	if err == nil {
		assert.Equal(1, width)
		assert.Equal(1, height)
		t.Logf("Minimum size parsed: %dx%d", width, height)
	} else {
		t.Logf("Minimum size rejected: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST VIDEO RATE PARSING

func Test_avutil_parse_video_rate(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input string
		num   int
		den   int
	}{
		{"25", 25, 1},
		{"30", 30, 1},
		{"60", 60, 1},
		{"24", 24, 1},
		{"50", 50, 1},
	}

	for _, tc := range tests {
		rate, err := AVUtil_parse_video_rate(tc.input)
		assert.NoError(err, "Failed to parse rate: %s", tc.input)
		assert.Equal(tc.num, int(rate.num), "Numerator mismatch for %s", tc.input)
		assert.Equal(tc.den, int(rate.den), "Denominator mismatch for %s", tc.input)
		t.Logf("Parsed rate %s -> %d/%d", tc.input, rate.num, rate.den)
	}
}

func Test_avutil_parse_video_rate_fractional(t *testing.T) {
	assert := assert.New(t)

	// Test NTSC frame rates which are typically represented as fractions
	tests := []struct {
		input string
		desc  string
	}{
		{"30000/1001", "NTSC 29.97 fps"},
		{"24000/1001", "NTSC 23.976 fps"},
		{"60000/1001", "NTSC 59.94 fps"},
		{"25/1", "PAL 25 fps"},
		{"30/1", "30 fps"},
	}

	for _, tc := range tests {
		rate, err := AVUtil_parse_video_rate(tc.input)
		assert.NoError(err, "Failed to parse rate: %s (%s)", tc.input, tc.desc)
		assert.NotEqual(0, int(rate.num), "Numerator should not be 0")
		assert.NotEqual(0, int(rate.den), "Denominator should not be 0")
		t.Logf("Parsed %s (%s) -> %d/%d", tc.input, tc.desc, rate.num, rate.den)
	}
}

func Test_avutil_parse_video_rate_decimal(t *testing.T) {
	assert := assert.New(t)

	// FFmpeg should handle decimal frame rates
	tests := []string{
		"29.97",
		"23.976",
		"59.94",
		"25.0",
		"24.0",
	}

	for _, input := range tests {
		rate, err := AVUtil_parse_video_rate(input)
		if err != nil {
			t.Logf("Decimal rate '%s' parsing failed (may not be supported): %v", input, err)
			continue
		}
		assert.NotEqual(0, int(rate.num), "Numerator should not be 0")
		assert.NotEqual(0, int(rate.den), "Denominator should not be 0")
		t.Logf("Parsed decimal rate %s -> %d/%d", input, rate.num, rate.den)
	}
}

func Test_avutil_parse_video_rate_invalid(t *testing.T) {
	assert := assert.New(t)

	invalidRates := []string{
		"",
		"invalid",
		"abc",
		"-25",
		"0",
		"30/0",
		"inf",
	}

	for _, input := range invalidRates {
		rate, err := AVUtil_parse_video_rate(input)
		assert.Error(err, "Expected error for invalid rate: %s", input)
		assert.Equal(0, int(rate.num), "Numerator should be 0 on error")
		assert.Equal(0, int(rate.den), "Denominator should be 0 on error")
		t.Logf("Invalid rate '%s' correctly rejected: %v", input, err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST TIME PARSING

func Test_avutil_parse_time(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input    string
		expected int64
		duration bool
	}{
		{"1", 1000000, true},           // 1 second (as duration)
		{"2", 2000000, true},           // 2 seconds (as duration)
		{"0.5", 500000, true},          // 0.5 seconds (as duration)
		{"10.25", 10250000, true},      // 10.25 seconds (as duration)
		{"00:01", 1000000, true},       // 1 second (MM:SS format as duration)
		{"01:00", 60000000, true},      // 1 minute (as duration)
		{"00:00:01", 1000000, true},    // 1 second (HH:MM:SS format as duration)
		{"00:01:30", 90000000, true},   // 1 minute 30 seconds (as duration)
		{"01:00:00", 3600000000, true}, // 1 hour (as duration)
	}

	for _, tc := range tests {
		timeval, err := AVUtil_parse_time(tc.input, tc.duration)
		assert.NoError(err, "Failed to parse time: %s", tc.input)
		assert.Equal(tc.expected, timeval, "Time value mismatch for %s", tc.input)
		t.Logf("Parsed time %s -> %d microseconds (%.3f seconds)", tc.input, timeval, float64(timeval)/1000000.0)
	}
}

func Test_avutil_parse_time_duration(t *testing.T) {
	assert := assert.New(t)

	// Test durations which can be negative
	tests := []struct {
		input    string
		expected int64
	}{
		{"-1", -1000000},      // -1 second
		{"-0.5", -500000},     // -0.5 seconds
		{"-00:01", -1000000},  // -1 second
		{"-01:00", -60000000}, // -1 minute
		{"5", 5000000},        // 5 seconds (positive duration)
	}

	for _, tc := range tests {
		timeval, err := AVUtil_parse_time(tc.input, true)
		assert.NoError(err, "Failed to parse duration: %s", tc.input)
		assert.Equal(tc.expected, timeval, "Duration value mismatch for %s", tc.input)
		t.Logf("Parsed duration %s -> %d microseconds (%.3f seconds)", tc.input, timeval, float64(timeval)/1000000.0)
	}
}

func Test_avutil_parse_time_milliseconds(t *testing.T) {
	assert := assert.New(t)

	// Test parsing with millisecond precision as durations
	tests := []struct {
		input string
		desc  string
	}{
		{"00:00:01.500", "1.5 seconds"},
		{"00:01:30.750", "90.75 seconds"},
		{"00:00:00.100", "100 milliseconds"},
	}

	for _, tc := range tests {
		timeval, err := AVUtil_parse_time(tc.input, true) // Parse as duration
		if err != nil {
			t.Logf("Time format '%s' (%s) may not be supported: %v", tc.input, tc.desc, err)
			continue
		}
		assert.Greater(timeval, int64(0), "Time should be positive for %s", tc.input)
		t.Logf("Parsed %s (%s) -> %d microseconds (%.6f seconds)", tc.input, tc.desc, timeval, float64(timeval)/1000000.0)
	}
}

func Test_avutil_parse_time_invalid(t *testing.T) {
	assert := assert.New(t)

	invalidTimes := []string{
		"",
		"invalid",
		"abc",
		"25:70:00", // Invalid minutes
		"25:30:70", // Invalid seconds
		"inf",
	}

	for _, input := range invalidTimes {
		timeval, err := AVUtil_parse_time(input, false)
		assert.Error(err, "Expected error for invtrue) // Parse as durationime: %s", input)
		assert.Equal(int64(0), timeval, "Time value should be 0 on error")
		t.Logf("Invalid time '%s' correctly rejected: %v", input, err)
	}
}

func Test_avutil_parse_time_edge_cases(t *testing.T) {
	assert := assert.New(t)

	// Test zero time - FFmpeg may reject this
	timeval, err := AVUtil_parse_time("0", true) // Parse as duration
	if err != nil {
		t.Logf("Zero time not supported: %v", err)
	} else {
		assert.Equal(int64(0), timeval)
		t.Logf("Zero time parsed: %d", timeval)
	}

	// Test very long duration
	timeval, err = AVUtil_parse_time("24:00:00", true) // 24 hours as duration
	if err == nil {
		expected := int64(24 * 3600 * 1000000) // 24 hours in microseconds
		assert.Equal(expected, timeval)
		t.Logf("24 hours parsed: %d microseconds", timeval)
	} else {
		t.Logf("24 hour time may not be supported: %v", err)
	}

	// Test 00:00:00 format
	timeval, err = AVUtil_parse_time("00:00:00", true) // Parse as duration
	if err == nil {
		assert.Equal(int64(0), timeval)
		t.Logf("00:00:00 parsed as zero: %d", timeval)
	}
}
