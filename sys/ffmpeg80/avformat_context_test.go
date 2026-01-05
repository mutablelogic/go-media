package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - ALLOCATION AND DEALLOCATION

func Test_AVFormat_alloc_context_001(t *testing.T) {
	assert := assert.New(t)

	ctx := AVFormat_alloc_context()
	assert.NotNil(ctx)

	AVFormat_free_context(ctx)
}

func Test_AVFormat_alloc_context_002(t *testing.T) {
	assert := assert.New(t)

	// Allocate multiple contexts
	contexts := make([]*AVFormatContext, 10)
	for i := 0; i < 10; i++ {
		contexts[i] = AVFormat_alloc_context()
		assert.NotNil(contexts[i])
	}

	// Free all contexts
	for _, ctx := range contexts {
		AVFormat_free_context(ctx)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - WITH ACTUAL FILES

func Test_AVFormatContext_properties_001(t *testing.T) {
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

	// Test basic properties
	duration := input.Duration()
	assert.Greater(duration, int64(0))
	t.Logf("Duration: %d", duration)

	startTime := input.StartTime()
	t.Logf("Start time: %d", startTime)

	bitRate := input.BitRate()
	t.Logf("Bit rate: %d", bitRate)

	filename := input.Filename()
	assert.Contains(filename, "sample.mp4")
	t.Logf("Filename: %s", filename)
}

func Test_AVFormatContext_streams_001(t *testing.T) {
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

	// Test NumStreams
	numStreams := input.NumStreams()
	assert.Greater(numStreams, uint(0))
	t.Logf("Number of streams: %d", numStreams)

	// Test Streams method
	streams := input.Streams()
	assert.Equal(int(numStreams), len(streams))

	for i, stream := range streams {
		assert.NotNil(stream)
		t.Logf("Stream %d: %p", i, stream)
	}
}

func Test_AVFormatContext_stream_access_001(t *testing.T) {
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

	// Test Stream method for valid indices
	for i := 0; i < int(input.NumStreams()); i++ {
		stream := input.Stream(i)
		assert.NotNil(stream)
		assert.Equal(i, stream.Index())
	}

	// Test Stream method with invalid indices
	assert.Nil(input.Stream(-1))
	assert.Nil(input.Stream(int(input.NumStreams())))
	assert.Nil(input.Stream(100))
}

func Test_AVFormatContext_input_format_001(t *testing.T) {
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

	// Test Input method
	inputFormat := input.Input()
	assert.NotNil(inputFormat)
	t.Logf("Input format: %s", inputFormat.Name())

	// Test Output method (should be nil for input contexts)
	outputFormat := input.Output()
	assert.Nil(outputFormat)
}

func Test_AVFormatContext_flags_001(t *testing.T) {
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

	// Test getting flags
	flags := input.Flags()
	t.Logf("Flags: %s", flags)

	// Test setting flags
	newFlags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	input.SetFlags(newFlags)
	assert.Equal(newFlags, input.Flags())
}

func Test_AVFormatContext_metadata_001(t *testing.T) {
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

	// Test getting metadata
	metadata := input.Metadata()
	assert.NotNil(metadata)

	// Try to get some entries
	entries := AVUtil_dict_entries(metadata)
	t.Logf("Metadata entries: %d", len(entries))

	for _, entry := range entries {
		t.Logf("Metadata: %s = %s", entry.Key(), entry.Value())
	}
}

func Test_AVFormatContext_multiple_files_001(t *testing.T) {
	assert := assert.New(t)

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

		// Test properties
		t.Logf("File: %s", filepath.Base(testFile))
		t.Logf("  Streams: %d", input.NumStreams())
		t.Logf("  Duration: %d", input.Duration())
		t.Logf("  Start time: %d", input.StartTime())
		t.Logf("  Bit rate: %d", input.BitRate())
		t.Logf("  Format: %s", input.Input().Name())

		AVFormat_close_input(input)
	}
}

func Test_AVFormatContext_pb_001(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	// Create custom IO context
	filereader, err := NewFileReader(testFile)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer filereader.Close()

	ioCtx := AVFormat_avio_alloc_context(20, false, filereader)
	assert.NotNil(ioCtx)
	defer AVFormat_avio_context_free(ioCtx)

	// Open for demuxing with custom IO
	input, err := AVFormat_open_reader(ioCtx, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_free_context(input)

	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Verify we can access streams
	assert.Greater(input.NumStreams(), uint(0))
	t.Logf("Successfully opened file with custom IO context")
}

func Test_AVFormatContext_duration_comparison_001(t *testing.T) {
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

	// Compare container duration with stream durations
	containerDuration := input.Duration()
	t.Logf("Container duration: %d", containerDuration)

	for i := 0; i < int(input.NumStreams()); i++ {
		stream := input.Stream(i)
		if stream != nil {
			// Note: Stream duration needs to be converted using time_base
			t.Logf("Stream %d index: %d", i, stream.Index())
		}
	}
}

func Test_AVFormatContext_probesize_001(t *testing.T) {
	assert := assert.New(t)

	// Test probe size getter/setter
	ctx := AVFormat_alloc_context()
	assert.NotNil(ctx)
	defer AVFormat_free_context(ctx)

	// Default probe size
	defaultProbe := ctx.ProbeSize()
	t.Logf("Default probe size: %d", defaultProbe)

	// Set custom probe size
	ctx.SetProbeSize(1024 * 1024)
	assert.Equal(int64(1024*1024), ctx.ProbeSize())

	// Test max analyze duration
	defaultAnalyze := ctx.MaxAnalyzeDuration()
	t.Logf("Default max analyze duration: %d", defaultAnalyze)

	// Set custom analyze duration
	ctx.SetMaxAnalyzeDuration(5000000)
	assert.Equal(int64(5000000), ctx.MaxAnalyzeDuration())
}

func Test_AVFormatContext_chapters_programs_001(t *testing.T) {
	assert := assert.New(t)

	// Open a test file
	input, err := AVFormat_open_url(filepath.Join("..", "..", "etc", "test", "sample.mp4"), nil, nil)
	assert.NoError(err)
	assert.NotNil(input)
	defer AVFormat_close_input(input)

	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Test chapters
	numChapters := input.NumChapters()
	t.Logf("Number of chapters: %d", numChapters)

	// Test programs
	numPrograms := input.NumPrograms()
	t.Logf("Number of programs: %d", numPrograms)

	// Test context flags
	ctxFlags := input.ContextFlags()
	t.Logf("Context flags: %d", ctxFlags)
}
