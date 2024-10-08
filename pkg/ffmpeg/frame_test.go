package ffmpeg_test

import (
	"testing"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
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
	assert.Equal(AUDIO, frame.Type())
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
	assert.Equal(VIDEO, frame.Type())
	assert.Equal(1280, frame.Width())
	assert.Equal(720, frame.Height())
	t.Log(frame)
}

func Test_frame_004(t *testing.T) {
	assert := assert.New(t)

	frame, err := ffmpeg.NewFrame(ffmpeg.VideoPar("rgba", "1280x720", 25))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	copy, err := frame.Copy()
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer copy.Close()

	assert.Equal(copy.Type(), frame.Type())
	assert.Equal(copy.PixelFormat(), frame.PixelFormat())
	assert.Equal(copy.Width(), frame.Width())
	assert.Equal(copy.Height(), frame.Height())
}

func Test_frame_005(t *testing.T) {
	assert := assert.New(t)

	frame, err := ffmpeg.NewFrame(ffmpeg.AudioPar("fltp", "stereo", 16000))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	t.Log(frame)
	data := make([]float32, frame.SampleRate())
	frame.SetFloat32(0, data)
	frame.SetFloat32(1, data)
	t.Log(frame)

	frame.SetFloat32(0, data[0:100])
	frame.SetFloat32(1, data[0:100])
	t.Log(frame)

}

func Test_frame_006(t *testing.T) {
	assert := assert.New(t)

	frame, err := ffmpeg.NewFrame(ffmpeg.AudioPar("fltp", "stereo", 16000))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	for i := 0; i < 10; i++ {
		frame.SetPts(int64(i))
		assert.Equal(int64(i), frame.Pts())
	}

	for i := 0; i < 10; i++ {
		frame.SetTs(float64(i))
		assert.Equal(float64(i), frame.Ts())
	}
}
