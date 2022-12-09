package audio_test

import (
	"fmt"
	"testing"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/audio"
)

func Test_swresample_000(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)
	ctx := mgr.NewContext()
	assert.NotNil(ctx)

	// Set source and desired output format
	assert.NoError(ctx.SetIn(AudioFormat{Rate: 48000, Format: SAMPLE_FORMAT_U8, Layout: CHANNEL_LAYOUT_STEREO}))
	assert.NoError(ctx.SetOut(AudioFormat{Rate: 44100, Format: SAMPLE_FORMAT_U8, Layout: CHANNEL_LAYOUT_STEREO}))

	// Convert
	in := make([]byte, 1024)
	assert.NoError(mgr.ConvertBytes(ctx, func(_ SWResampleContext, out []byte) ([]byte, error) {
		fmt.Println("out=", out)
		return in, nil
	}))

	// Close
	assert.NoError(mgr.Close())
}
