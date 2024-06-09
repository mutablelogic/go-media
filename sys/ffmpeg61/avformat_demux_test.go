package ffmpeg_test

import (
	"fmt"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avformat_demux_001(t *testing.T) {
	assert := assert.New(t)

	// Populate the data
	for i := 0; i < 100; i++ {
		data.WriteString(fmt.Sprintf("%v: hello, world\n", i))
	}

	// Create the context
	ctx := AVFormat_avio_alloc_context(20, false, new(reader))
	assert.NotNil(ctx)
	defer AVFormat_avio_context_free(ctx)

	// Open for demuxing
	input, err := AVFormat_open_reader(ctx, nil, nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer AVFormat_free_context(input)
}
