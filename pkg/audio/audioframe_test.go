package audio_test

import (
	"testing"
	"time"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/audio"
)

func Test_audioframe_000(t *testing.T) {
	assert := assert.New(t)
	frame, err := NewAudioFrame(AudioFormat{Rate: 48000, Format: SAMPLE_FORMAT_U8}, 0)
	assert.NoError(err)
	assert.NotNil(frame)
	t.Log(frame)
	assert.NoError(frame.Close())
}

func Test_audioframe_001(t *testing.T) {
	assert := assert.New(t)
	frame, err := NewAudioFrame(AudioFormat{Rate: 48000, Format: SAMPLE_FORMAT_U8, Layout: CHANNEL_LAYOUT_STEREO}, time.Second)
	assert.NoError(err)
	assert.NotNil(frame)
	t.Log(frame)
	assert.NoError(frame.Close())
}

func Test_audioframe_002(t *testing.T) {
	assert := assert.New(t)
	frame, err := NewAudioFramePlanar(AudioFormat{Rate: 48000, Format: SAMPLE_FORMAT_U8, Layout: CHANNEL_LAYOUT_STEREO}, time.Second, true)
	assert.NoError(err)
	assert.NotNil(frame)
	t.Log(frame)
	assert.NoError(frame.Close())
}

func Test_audioframe_003(t *testing.T) {
	assert := assert.New(t)
	frame, err := NewAudioFramePlanar(AudioFormat{Rate: 48000, Format: SAMPLE_FORMAT_U8, Layout: CHANNEL_LAYOUT_STEREO}, time.Second, true)
	assert.NoError(err)
	assert.NotNil(frame)
	t.Log(frame)
	for i, ch := range frame.Channels() {
		t.Log("channel", ch, "size=", len(frame.Bytes(i)), "bytes")
	}
	assert.NoError(frame.Close())
}
