package ffmpeg_test

import (
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avutil_samples_000(t *testing.T) {
	assert := assert.New(t)
	num_channels := 6
	num_samples := 1
	format := AV_SAMPLE_FMT_U8P
	data, err := AVUtil_samples_alloc(num_samples, num_channels, format, true)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(data)
	AVUtil_samples_set_silence(data, 0, num_samples)

	for plane := 0; plane < data.NumPlanes(); plane++ {
		assert.NotNil(data.Bytes(plane))
	}

	AVUtil_samples_free(data)
}
