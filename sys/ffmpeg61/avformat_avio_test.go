package ffmpeg_test

import (
	"bytes"
	"fmt"
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
	"github.com/stretchr/testify/assert"
)

var (
	// Create some data to read from
	data = new(bytes.Buffer)
)

func reader(buf []byte) int {
	n, err := data.Read(buf)
	if err != nil {
		return AVERROR_EOF
	} else {
		return n
	}
}

func Test_avio_001(t *testing.T) {
	assert := assert.New(t)

	// Populate the data
	for i := 0; i < 100; i++ {
		data.WriteString(fmt.Sprintf("%v: hello, world\n", i))
	}

	// Create the context
	ctx := AVFormat_avio_alloc_context(20, false, reader, nil, nil)
	assert.NotNil(ctx)

	// Read the data
	var buf [100]byte
	for {
		n := AVFormat_avio_read(ctx, buf[:])
		if n == AVERROR_EOF {
			break
		}
		fmt.Println("N=", n, string(buf[:n]))
	}

	// Free the context
	AVFormat_avio_context_free(ctx)
}
