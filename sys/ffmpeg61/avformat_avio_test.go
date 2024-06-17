package ffmpeg_test

import (
	"bytes"
	"fmt"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

var (
	// Create some data to read from
	data = new(bytes.Buffer)
)

type reader struct{}

func (r *reader) Reader(buf []byte) int {
	n, err := data.Read(buf)
	if err != nil {
		return AVERROR_EOF
	} else {
		return n
	}
}

func (r *reader) Writer([]byte) int {
	return AVERROR_EOF
}

func (r *reader) Seeker(int64, int) int64 {
	return -1
}

func Test_avio_001(t *testing.T) {
	assert := assert.New(t)

	// Populate the data
	for i := 0; i < 100; i++ {
		data.WriteString(fmt.Sprintf("%v: hello, world\n", i))
	}

	// Create the context
	ctx := AVFormat_avio_alloc_context(20, false, new(reader))
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
