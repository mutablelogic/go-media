package ffmpeg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - WITH ACTUAL FILES

func Test_avformat_stream_001(t *testing.T) {
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

	// Check that we have at least one stream
	numStreams := input.NumStreams()
	assert.Greater(numStreams, uint(0))

	t.Logf("Found %d streams in MP4 file", numStreams)
}

func Test_avformat_stream_002(t *testing.T) {
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

	// Access each stream and verify properties
	for i := uint(0); i < input.NumStreams(); i++ {
		stream := input.Stream(int(i))
		assert.NotNil(stream)
		assert.Equal(int(i), stream.Index())

		// Test property accessors
		_ = stream.Id()
		_ = stream.TimeBase()
		_ = stream.Disposition()
		codecPar := stream.CodecPar()
		assert.NotNil(codecPar)

		t.Logf("Stream %d: Index=%d, Id=%d", i, stream.Index(), stream.Id())
	}
}

func Test_avformat_stream_003(t *testing.T) {
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

	// Test JSON marshaling for each stream
	for i := uint(0); i < input.NumStreams(); i++ {
		stream := input.Stream(int(i))
		data, err := json.Marshal(stream)
		assert.NoError(err)
		assert.NotEmpty(data)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		assert.NoError(err)
		assert.Contains(result, "index")

		t.Logf("Stream %d JSON: %s", i, string(data))
	}
}

func Test_avformat_stream_004(t *testing.T) {
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

	// Test String() method
	for i := uint(0); i < input.NumStreams(); i++ {
		stream := input.Stream(int(i))
		str := stream.String()
		assert.NotEmpty(str)
		assert.Contains(str, "index")

		t.Logf("Stream %d String:\n%s", i, str)
	}
}

func Test_AVFormat_find_best_stream_001(t *testing.T) {
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

	// Find best video stream
	videoIdx, codec, err := AVFormat_find_best_stream(input, AVMEDIA_TYPE_VIDEO, -1, -1)
	if err == nil {
		assert.GreaterOrEqual(videoIdx, 0)
		t.Logf("Found video stream at index %d, codec: %v", videoIdx, codec)
	} else {
		t.Log("No video stream found (acceptable for audio-only files)")
	}
}

func Test_AVFormat_find_best_stream_002(t *testing.T) {
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

	// Find best audio stream
	audioIdx, codec, err := AVFormat_find_best_stream(input, AVMEDIA_TYPE_AUDIO, -1, -1)
	if err == nil {
		assert.GreaterOrEqual(audioIdx, 0)
		t.Logf("Found audio stream at index %d, codec: %v", audioIdx, codec)
	}
}

func Test_AVFormat_find_best_stream_003(t *testing.T) {
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

	// Find all stream types
	mediaTypes := []struct {
		name string
		mt   AVMediaType
	}{
		{"video", AVMEDIA_TYPE_VIDEO},
		{"audio", AVMEDIA_TYPE_AUDIO},
		{"subtitle", AVMEDIA_TYPE_SUBTITLE},
		{"data", AVMEDIA_TYPE_DATA},
	}

	for _, mt := range mediaTypes {
		idx, _, err := AVFormat_find_best_stream(input, mt.mt, -1, -1)
		if err == nil {
			t.Logf("Found %s stream at index %d", mt.name, idx)
		} else {
			t.Logf("No %s stream found", mt.name)
		}
	}
}

func Test_avformat_stream_disposition_001(t *testing.T) {
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

	// Check stream dispositions
	for i := uint(0); i < input.NumStreams(); i++ {
		stream := input.Stream(int(i))
		disposition := stream.Disposition()
		t.Logf("Stream %d disposition: %s", i, disposition)

		// Test AttachedPic
		attachedPic := stream.AttachedPic()
		if attachedPic != nil {
			t.Logf("Stream %d has attached picture", i)
		}
	}
}

func Test_avformat_stream_timebase_001(t *testing.T) {
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

	// Check stream time bases
	for i := uint(0); i < input.NumStreams(); i++ {
		stream := input.Stream(int(i))
		timeBase := stream.TimeBase()
		t.Logf("Stream %d time_base: %v", i, timeBase)
	}
}
