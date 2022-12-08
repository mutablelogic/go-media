package ffmpeg_test

import (
	"testing"

	// Pacakge imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
	assert "github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_swr_000(t *testing.T) {
	t.Log("SWR_version=", ffmpeg.SWR_version())
}

func Test_swr_001(t *testing.T) {
	t.Log("SWR_version=", ffmpeg.SWR_configuration())
}

func Test_swr_002(t *testing.T) {
	t.Log("SWR_license=", ffmpeg.SWR_license())
}

func Test_swr_003(t *testing.T) {
	assert := assert.New(t)
	swr := ffmpeg.SWR_alloc()
	assert.NotNil(swr)
	t.Log(swr)
	swr.SWR_free()
}
