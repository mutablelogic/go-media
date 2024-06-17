package media_test

import (
	"os"
	"testing"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/media"
)

func Test_map_000(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	defer mgr.Close()

	// Open input file
	media, err := mgr.OpenFile(SAMPLE_MP4, nil)
	assert.NoError(err)
	assert.NotNil(media)

	// Create a map for audio
	mapping, err := mgr.Map(media, MEDIA_FLAG_AUDIO)
	assert.NoError(err)
	assert.NotNil(mapping)

	// Resample audio
	assert.NoError(mapping.Resample(AudioFormat{
		Rate:   44100,
		Format: SAMPLE_FORMAT_U8,
	}, mapping.Streams(MEDIA_FLAG_AUDIO)[0]))

	// Print out the map
	mapping.PrintMap(os.Stdout)
}

func Test_map_001(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	defer mgr.Close()

	// Open input file
	media, err := mgr.OpenFile(SAMPLE_MP4, nil)
	assert.NoError(err)
	assert.NotNil(media)

	// Create a map for audio
	mapping, err := mgr.Map(media, MEDIA_FLAG_AUDIO)
	assert.NoError(err)
	assert.NotNil(mapping)

	// Resample audio
	assert.NoError(mapping.Resample(AudioFormat{
		Rate:   11025,
		Format: SAMPLE_FORMAT_U8,
		Layout: CHANNEL_LAYOUT_MONO,
	}, media.Streams()[1]))

	// Print out the map
	mapping.PrintMap(os.Stdout)
}
