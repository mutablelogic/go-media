package media_test

import (
	"testing"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	//. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/media"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
)

func Test_metadata_000(t *testing.T) {
	var dict *ffmpeg.AVDictionary
	assert := assert.New(t)
	// Create a new metadata object
	metadata := NewMetadata(&dict)
	assert.NotNil(metadata)
	t.Log(metadata)
	//err := metadata.Close()
	//assert.NoError(err)
}

func Test_metadata_001(t *testing.T) {
	var dict *ffmpeg.AVDictionary
	assert := assert.New(t)
	// Create a new metadata object
	metadata := NewMetadata(&dict)
	assert.NotNil(metadata)
	err := metadata.Set("foo", "bar")
	assert.NoError(err)
	t.Log(metadata)
	assert.Equal(1, metadata.Count())
	err = metadata.Set("foo", "bar")
	assert.NoError(err)
	assert.Equal(1, metadata.Count())
	err = metadata.Set("foo", nil)
	assert.NoError(err)
	assert.Equal(0, metadata.Count())
	//err = metadata.Close()
	//assert.NoError(err)
}

/*
func Test_metadata_002(t *testing.T) {
	assert := assert.New(t)
	// Create a new metadata object
	metadata := NewMetadata(nil)
	assert.NotNil(metadata)
	err := metadata.Set("foo", "bar")
	assert.NoError(err)
	err = metadata.Set("foo", nil)
	assert.NoError(err)
	assert.Equal(0, metadata.Count())
	t.Log(metadata)
	err = metadata.Close()
	assert.NoError(err)
}
*/
