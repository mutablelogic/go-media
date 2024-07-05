package ffmpeg_test

import (
	"testing"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
	assert "github.com/stretchr/testify/assert"
)

func Test_resampler_001(t *testing.T) {
	assert := assert.New(t)

	// Sine wave generator
	audio, err := generator.NewSine(2000, 10, ffmpeg.AudioPar("fltp", "mono", 22050))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer audio.Close()

	// Create a rescaler
	resampler, err := ffmpeg.NewResampler(ffmpeg.AudioPar("s16", "stereo", 44100), false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer resampler.Close()

	// Rescale ten frames
	for i := 0; i < 10; i++ {
		src := audio.Frame()
		if !assert.NotNil(src) {
			t.FailNow()
		}

		// Rescale the frame
		dest, err := resampler.Frame(src)
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Display information
		t.Log(src, "=>", dest)
	}

	// Flush
	for {
		dest, err := resampler.Frame(nil)
		if !assert.NoError(err) {
			t.FailNow()
		}
		t.Log(" =>", dest)
		if dest == nil {
			break
		}
	}
}
