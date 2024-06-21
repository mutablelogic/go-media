package media_test

import (
	// Import namespaces
	"context"
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func Test_decoder_001(t *testing.T) {
	assert := assert.New(t)

	manager := NewManager()
	media, err := manager.Open("./etc/test/sample.mp4", nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer media.Close()

	decoder, err := media.Decoder(func(stream Stream) (Parameters, error) {
		// Copy parameters from the stream
		return stream.Parameters(), nil
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Demuliplex the stream
	assert.NoError(decoder.Demux(context.Background(), nil))
}
