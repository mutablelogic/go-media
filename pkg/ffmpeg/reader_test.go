package ffmpeg_test

import (
	"io"
	"os"
	"testing"

	"github.com/mutablelogic/go-media/pkg/ffmpeg"
	"github.com/stretchr/testify/assert"
)

func Test_reader_00(t *testing.T) {
	assert := assert.New(t)

	// Open a stream
	r, err := os.Open("reader.go")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	// Create an ffmpeg reader
	r2 := ffmpeg.NewReader(r)
	assert.NotNil(r2)
	defer r2.Close()

	// Read the data
	var buf [100]byte
	for {
		n, err := r2.Read(buf[:])
		if err == io.EOF {
			break
		}
		assert.NoError(err)
		assert.NotEqual(0, n)
		t.Log(string(buf[:n]))
	}
}
