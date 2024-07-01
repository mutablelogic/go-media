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

func Test_frame_002(t *testing.T) {
	assert := assert.New(t)

	frame, err := ffmpeg.NewFrame(ffmpeg.AudioPar("s16", "stereo", 44100))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()
	assert.Equal(ffmpeg.AUDIO, frame.Type())
	assert.Equal(44100, frame.SampleRate())
	t.Log(frame)
}

func Test_frame_003(t *testing.T) {
	assert := assert.New(t)

	frame, err := ffmpeg.NewFrame(ffmpeg.VideoPar("rgba", "1280x720", 25))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()
	assert.Equal(ffmpeg.VIDEO, frame.Type())
	assert.Equal(1280, frame.Width())
	assert.Equal(720, frame.Height())
	t.Log(frame)
}
