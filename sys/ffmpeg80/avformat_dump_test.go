package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - BASIC FUNCTIONALITY

func Test_avformat_dump_001(t *testing.T) {
	// Test that dump_format doesn't crash with nil context
	// This should not panic but may print an error to stderr
	AVFormat_dump_format(nil, 0, "test.mp4")

	t.Log("AVFormat_dump_format with nil context completed without panic")
}

func Test_avformat_dump_002(t *testing.T) {
	// Test with various stream indices and nil context
	// These should not panic
	indices := []int{-1, 0, 1, 100}
	for _, idx := range indices {
		AVFormat_dump_format(nil, idx, "test.mp4")
		t.Logf("AVFormat_dump_format with nil context and stream index %d completed", idx)
	}
}

func Test_avformat_dump_003(t *testing.T) {
	// Test with empty filename
	AVFormat_dump_format(nil, 0, "")

	t.Log("AVFormat_dump_format with empty filename completed without panic")
}

func Test_avformat_dump_004(t *testing.T) {
	// Test with various filenames
	filenames := []string{
		"test.mp4",
		"test.avi",
		"test.mkv",
		"/path/to/file.mp4",
		"rtmp://server/stream",
		"http://example.com/video.mp4",
	}

	for _, filename := range filenames {
		AVFormat_dump_format(nil, 0, filename)
		t.Logf("AVFormat_dump_format with filename %s completed", filename)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - WITH ACTUAL FILES

func Test_avformat_dump_005(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	// Open input file
	input, err := AVFormat_open_url(testFile, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	// Find stream information
	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Dump format information (will print to stderr)
	AVFormat_dump_format(input, 0, testFile)

	t.Log("Successfully dumped format for MP4 file")
}

func Test_avformat_dump_006(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp3")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	input, err := AVFormat_open_url(testFile, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Dump with stream index -1 (all streams)
	AVFormat_dump_format(input, -1, testFile)

	t.Log("Successfully dumped all streams for MP3 file")
}

func Test_avformat_dump_007(t *testing.T) {
	assert := assert.New(t)

	// Test with multiple file formats
	testFiles := []string{
		filepath.Join("..", "..", "etc", "test", "sample.mp4"),
		filepath.Join("..", "..", "etc", "test", "sample.mp3"),
		filepath.Join("..", "..", "etc", "test", "jfk.wav"),
	}

	for _, testFile := range testFiles {
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Logf("Skipping %s: file not available", filepath.Base(testFile))
			continue
		}

		input, err := AVFormat_open_url(testFile, nil, nil)
		if !assert.NoError(err) {
			continue
		}

		if err := AVFormat_find_stream_info(input, nil); !assert.NoError(err) {
			AVFormat_close_input(input)
			continue
		}

		// Dump each stream individually
		for i := 0; i < int(input.NumStreams()); i++ {
			AVFormat_dump_format(input, i, testFile)
			t.Logf("Dumped stream %d for %s", i, filepath.Base(testFile))
		}

		AVFormat_close_input(input)
	}
}
