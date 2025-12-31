package ffmpeg_test

import (
	"testing"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_par_001(t *testing.T) {
	assert := assert.New(t)

	par, err := ffmpeg.NewAudioPar("fltp", "mono", 22050)
	if !assert.NoError(err) {
		t.FailNow()
	}
	t.Log(par)
}

func Test_par_002(t *testing.T) {
	assert := assert.New(t)

	par, err := ffmpeg.NewVideoPar("yuv420p", "1280x720", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}
	t.Log(par)
}
