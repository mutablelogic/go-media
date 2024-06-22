package media_test

import (
	// Import namespaces
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func Test_decoder_001(t *testing.T) {
	// Decode packets
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
	// Decode video frames
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

func Test_decoder_003(t *testing.T) {
	// Decode video frames and resize them
	assert := assert.New(t)

	// Open the file
	manager := NewManager()
	media, err := manager.Open("./etc/test/sample.mp4", nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer media.Close()

	// Initialize
	n := 0
	// tmpdir := t.TempDir()
	tmpdir, err := os.MkdirTemp("", "media_test")
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Create a decoder which resizes the video frames
	decoder, err := media.Decoder(func(stream Stream) (Parameters, error) {
		// Rescale the video
		if stream.Type() == VIDEO {
			return manager.VideoParameters(640, 480, "yuv420p")
		}
		// Ignore other streams
		return nil, nil
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// This is the function which processes the frames
	framefn := func(frame Frame) error {
		// Create an output file
		filename := filepath.Join(tmpdir, fmt.Sprintf("frame%03d.jpg", n))
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()

		// Decode the frame into an image, save as JPEG
		if image, err := frame.Image(); err != nil {
			return err
		} else if err := jpeg.Encode(w, image, nil); err != nil {
			return err
		} else {
			t.Logf("Frame %d: %dx%d (%q) => %s", n, frame.Width(), frame.Height(), frame.PixelFormat(), filename)
			n++
		}

		// Stop after 10 frames
		if n >= 10 {
			return io.EOF
		} else {
			return nil
		}
	}

	// Finally, this is where we actually decode frames from the stream
	assert.NoError(decoder.Decode(context.Background(), framefn))
}

func Test_decoder_004(t *testing.T) {
	// Decode audio frames
	assert := assert.New(t)

	// Open the file
	manager := NewManager()
	media, err := manager.Open("./etc/test/sample.mp4", nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer media.Close()

	// Create a decoder to decompress the audio
	decoder, err := media.Decoder(func(stream Stream) (Parameters, error) {
		// Audio - pass through
		if stream.Type() == AUDIO {
			return stream.Parameters(), nil
		}
		// Ignore other streams
		return nil, nil
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// This is the function which processes the audio frames
	framefn := func(frame Frame) error {
		t.Log(frame)
		return nil
	}

	// Finally, this is where we actually decode frames from the stream
	assert.NoError(decoder.Decode(context.Background(), framefn))
}
