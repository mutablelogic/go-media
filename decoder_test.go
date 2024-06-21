package media_test

import (
	// Import namespaces
	"context"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
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

	// Packet function
	packetfn := func(packet Packet) error {
		// Null provided when flushing
		t.Log(packet)
		return nil
	}

	// Demuliplex the stream
	assert.NoError(decoder.Demux(context.Background(), packetfn))
}

func Test_decoder_002(t *testing.T) {
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

	// Frame function
	n := 0
	tmpdir := t.TempDir()
	if !assert.NoError(err) {
		t.SkipNow()
	}
	framefn := func(frame Frame) error {
		if frame.Type() != VIDEO {
			return nil
		}
		filename := filepath.Join(tmpdir, fmt.Sprintf("frame%03d.jpg", n))
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		if image, err := frame.Image(); err != nil {
			return err
		} else if err := jpeg.Encode(w, image, nil); err != nil {
			return err
		} else {
			t.Logf("Frame %d: %dx%d => %s", n, frame.Width(), frame.Height(), filename)
			n++
		}
		return nil
	}

	// decode frames from the stream
	assert.NoError(decoder.Decode(context.Background(), framefn))
}
