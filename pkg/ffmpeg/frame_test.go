package ffmpeg_test

import (
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_frame_001(t *testing.T) {
	assert := assert.New(t)

	frame, err := ffmpeg.NewFrame(nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()
	t.Log(frame)
}
