package ffmpeg_test

import (
	"os"
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_reader_001(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := ffmpeg.Open("../../etc/test/sample.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	t.Log(r)
}

func Test_reader_002(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := os.Open("../../etc/test/sample.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	media, err := ffmpeg.NewReader(r)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer media.Close()

	t.Log(media)
}
