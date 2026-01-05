package ffmpeg_test

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_reader_001(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := ffmpeg.Open("../../etc/test/sample.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	t.Log(r)
}

func Test_reader_002(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := os.Open("../../etc/test/sample.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	media, err := ffmpeg.NewReader(r)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer media.Close()

	t.Log(media)
}

func Test_reader_003(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := os.Open("../../etc/test/sample.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	media, err := ffmpeg.NewReader(r)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer media.Close()

	framefn := func(stream int, frame *ffmpeg.Frame) error {
		// Receive a nil frame at the end of each packet
		if frame != nil {
			t.Logf("Frame %v[%d] => %v", frame.Type(), stream, time.Duration(frame.Ts()*float64(time.Second)).Truncate(time.Millisecond))
		}
		return nil
	}

	if err := media.Decode(context.Background(), nil, framefn); !assert.NoError(err) {
		t.FailNow()
	}
}

func Test_reader_004(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := os.Open("../../etc/test/sample.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	media, err := ffmpeg.NewReader(r)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer media.Close()

	// Map function
	mapfn := func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		t.Logf("Stream %v[%d] => %v", par.Type(), stream, par)
		return par, nil
	}

	// Frame function
	framefn := func(stream int, frame *ffmpeg.Frame) error {
		t.Logf("Frame %v[%d] => %v", frame.Type(), stream, time.Duration(frame.Ts()*float64(time.Second)).Truncate(time.Millisecond))
		return nil
	}

	if err := media.Decode(context.Background(), mapfn, framefn); !assert.NoError(err) {
		t.FailNow()
	}
}

func Test_reader_005(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := os.Open("../../etc/test/sample.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	input, err := ffmpeg.NewReader(r)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer input.Close()

	// Map function - only video streams
	mapfn := func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.VIDEO {
			t.Logf("Stream %v[%d] => %v", par.Type(), stream, par)
			return par, nil
		}
		return nil, nil
	}

	// Frame function - save 10 thumbnails
	tmp, err := os.MkdirTemp("", t.Name())
	if !assert.NoError(err) {
		t.FailNow()
	}
	framefn := func(stream int, frame *ffmpeg.Frame) error {
		if frame.Ts() > 1.0 {
			return io.EOF
		}

		// Create the file
		w, err := os.Create(filepath.Join(tmp, fmt.Sprintf("frame-%d.jpg", frame.Pts())))
		if err != nil {
			return err
		}
		defer w.Close()

		// Get the image
		image, err := frame.Image()
		if err != nil {
			return err
		}
		if err := jpeg.Encode(w, image, nil); err != nil {
			return err
		}
		t.Log("Written frame", w.Name())
		return nil
	}

	if err := input.Decode(context.Background(), mapfn, framefn); !assert.NoError(err) {
		t.FailNow()
	}
}

func Test_reader_006(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	r, err := os.Open("../../etc/test/jfk.wav")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	input, err := ffmpeg.NewReader(r)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer input.Close()

	// Map function - only audio streams
	mapfn := func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.AUDIO {
			t.Logf("Stream %v[%d] => %v", par.Type(), stream, par)
			return par, nil
		}
		return nil, nil
	}

	framefn := func(stream int, frame *ffmpeg.Frame) error {
		t.Log("Got frame", frame)
		return nil
	}

	if err := input.Decode(context.Background(), mapfn, framefn); !assert.NoError(err) {
		t.FailNow()
	}
}
