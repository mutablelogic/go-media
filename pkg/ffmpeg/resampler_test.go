package ffmpeg_test

import (
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
	assert "github.com/stretchr/testify/assert"
)

func Test_resampler_001(t *testing.T) {
	assert := assert.New(t)

	// Sine wave generator
	audio, err := generator.NewSine(2000, 10, ffmpeg.AudioPar("fltp", "mono", 44100))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer audio.Close()

	// Create a rescaler
	resampler, err := ffmpeg.NewResampler(ff.AV_SAMPLE_FMT_S16P, ffmpeg.OptChannels(2))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer resampler.Close()

	// Rescale ten frames
	for i := 0; i < 10; i++ {
		src := audio.Frame().(*ffmpeg.Frame)
		if !assert.NotNil(src) {
			t.FailNow()
		}

		// Rescale the frame
		dest, err := resampler.Frame(src.AVFrame())
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Display information
		t.Log(src, "=>", dest)
	}

	// Flush
	dest, err := resampler.Frame(nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	t.Log("FLUSH =>", dest)
}
