package avcodec_test

import (
	"testing"

	// Packages
	media "github.com/mutablelogic/go-media"
	avcodec "github.com/mutablelogic/go-media/pkg/avcodec"
	assert "github.com/stretchr/testify/assert"
)

func Test_codec_001(t *testing.T) {
	assert := assert.New(t)

	codecs := avcodec.Codecs(media.OUTPUT | media.AUDIO)
	assert.NotNil(codecs)
	for _, meta := range codecs {
		assert.NotNil(meta)
		codec, err := avcodec.NewEncoder(meta.Key())
		defer codec.Close()

		if assert.NoError(err) {
			assert.NotNil(codec)
			assert.Equal(media.AUDIO, codec.Type())
			t.Log(codec)
		}
	}
}

func Test_codec_002(t *testing.T) {
	assert := assert.New(t)

	codecs := avcodec.Codecs(media.OUTPUT | media.VIDEO)
	assert.NotNil(codecs)
	for _, meta := range codecs {
		assert.NotNil(meta)
		codec, err := avcodec.NewEncoder(meta.Key())
		defer codec.Close()

		if assert.NoError(err) {
			assert.NotNil(codec)
			assert.Equal(media.VIDEO, codec.Type())
			t.Log(codec)
		}
	}
}
