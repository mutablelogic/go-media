package audio_test

import (
	"testing"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/audio"
)

func Test_audioframe_000(t *testing.T) {
	assert := assert.New(t)
	frame, err := NewAudioFrame(AudioFormat{Rate: 48000, Format: SAMPLE_FORMAT_U8, Layout: CHANNEL_LAYOUT_STEREO}, 0, false)
	assert.NoError(err)
	assert.NotNil(frame)
	t.Log(frame)
}
