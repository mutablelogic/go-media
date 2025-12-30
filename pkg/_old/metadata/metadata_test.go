package metadata_test

import (
	"os"
	"testing"

	// Packages
	metadata "github.com/mutablelogic/go-media/pkg/metadata"
	assert "github.com/stretchr/testify/assert"
)

func Test_metadata_001(t *testing.T) {
	assert := assert.New(t)

	// Create a metadata
	metadata := metadata.New("test", "test")
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
	meta := metadata.New(metadata.MetaArtwork, data)
	if !assert.NotNil(meta) {
		t.FailNow()
	}
	assert.Equal(metadata.MetaArtwork, meta.Key())
	assert.Equal("image/png", meta.Value())
	assert.Equal(data, meta.Bytes())

	image := meta.Image()
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
	meta := metadata.New(metadata.MetaArtwork, data)
	if !assert.NotNil(meta) {
		t.FailNow()
	}
	assert.Equal(metadata.MetaArtwork, meta.Key())
	assert.Equal("image/jpeg", meta.Value())
	assert.Equal(data, meta.Bytes())

	image := meta.Image()
	if !assert.NotNil(image) {
		t.FailNow()
	}
}
