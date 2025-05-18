package ffmpeg_test

import (
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avformat_dump_001(t *testing.T) {
	assert := assert.New(t)

	// Open input file
	input, err := AVFormat_open_url(TEST_MP4_FILE, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	// Fine stream information
	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Dump the input format
	AVFormat_dump_format(input, 0, TEST_MP4_FILE)
}
