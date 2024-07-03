package ffmpeg_test

import (
	"os"
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_metadata_001(t *testing.T) {
	assert := assert.New(t)

	// Create a metadata
	metadata := ffmpeg.NewMetadata("test", "test")
	if !assert.NotNil(metadata) {
		t.FailNow()
	}
	assert.Equal("test", metadata.Key())
	assert.Equal("test", metadata.Value())
}
func Test_metadata_002(t *testing.T) {
	assert := assert.New(t)

	data, err := os.ReadFile("../../etc/test/sample.png")
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create a metadata
	metadata := ffmpeg.NewMetadata(ffmpeg.MetaArtwork, data)
	if !assert.NotNil(metadata) {
		t.FailNow()
	}
	assert.Equal(ffmpeg.MetaArtwork, metadata.Key())
	assert.Equal("image/png", metadata.Value())
	assert.Equal(data, metadata.Bytes())

	image := metadata.Image()
	if !assert.NotNil(image) {
		t.FailNow()
	}
}

func Test_metadata_003(t *testing.T) {
	assert := assert.New(t)

	data, err := os.ReadFile("../../etc/test/sample.jpg")
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create a metadata
	metadata := ffmpeg.NewMetadata(ffmpeg.MetaArtwork, data)
	if !assert.NotNil(metadata) {
		t.FailNow()
	}
	assert.Equal(ffmpeg.MetaArtwork, metadata.Key())
	assert.Equal("image/jpeg", metadata.Value())
	assert.Equal(data, metadata.Bytes())

	image := metadata.Image()
	if !assert.NotNil(image) {
		t.FailNow()
	}
}
