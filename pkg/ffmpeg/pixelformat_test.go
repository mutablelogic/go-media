package ffmpeg_test

import (
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_pixelformat_001(t *testing.T) {
	assert := assert.New(t)

	manager, err := ffmpeg.NewManager()
	if !assert.NoError(err) {
		t.FailNow()
	}

	for _, format := range manager.PixelFormats() {
		t.Logf("%v", format)
	}
}
