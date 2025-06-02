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
		codec, err := avcodec.NewEncoder(meta.Key(), avcodec.WithSampleRate(22050))
		if assert.NoError(err) {
			assert.NotNil(codec)
			assert.Equal(media.OUTPUT|media.AUDIO, codec.Type())
			t.Log(codec)
			assert.NoError(codec.Close())
		}
	}
}

func Test_codec_002(t *testing.T) {
	assert := assert.New(t)

	codecs := avcodec.Codecs(media.OUTPUT | media.VIDEO)
	assert.NotNil(codecs)
	for _, meta := range codecs {
		assert.NotNil(meta)
		codec, err := avcodec.NewEncoder(meta.Key(), avcodec.WithFrameRate(1, 25), avcodec.WithFrameSize("hd720"))
		if assert.NoError(err) {
			assert.NotNil(codec)
			assert.Equal(media.OUTPUT|media.VIDEO, codec.Type())
			t.Log(codec)
			assert.NoError(codec.Close())
		}
	}
}

func Test_codec_003(t *testing.T) {
	assert := assert.New(t)

	codecs := avcodec.Codecs(media.INPUT | media.AUDIO)
	assert.NotNil(codecs)
	for _, meta := range codecs {
		assert.NotNil(meta)
		codec, err := avcodec.NewDecoder(meta.Key())
		if assert.NoError(err) {
			assert.NotNil(codec)
			assert.Equal(media.INPUT|media.AUDIO, codec.Type())
			t.Log(codec)
			assert.NoError(codec.Close())
		}
	}
}

func Test_codec_004(t *testing.T) {
	assert := assert.New(t)

	codecs := avcodec.Codecs(media.INPUT | media.VIDEO)
	assert.NotNil(codecs)
	for _, meta := range codecs {
		assert.NotNil(meta)
		codec, err := avcodec.NewDecoder(meta.Key(), avcodec.WithFrameSize("hd720"))
		if assert.NoError(err) {
			assert.NotNil(codec)
			assert.Equal(media.INPUT|media.VIDEO, codec.Type())
			t.Log(codec)
			assert.NoError(codec.Close())
		}
	}
}
