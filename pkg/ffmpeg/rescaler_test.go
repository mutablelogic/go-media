package ffmpeg_test

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
	assert "github.com/stretchr/testify/assert"
)

func Test_rescaler_001(t *testing.T) {
	assert := assert.New(t)

	// Create an image generator
	image, err := generator.NewYUV420P(ffmpeg.VideoPar("yuv420p", "1280x720", 25))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer image.Close()

	// Create a rescaler
	rescaler, err := ffmpeg.NewRescaler(ffmpeg.VideoPar("rgb24", "1024x768", 25), false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer rescaler.Close()

	// Rescale ten frames
	for i := 0; i < 10; i++ {
		src := image.Frame()
		if !assert.NotNil(src) {
			t.FailNow()
		}

		// Rescale the frame
		dest, err := rescaler.Frame(src)
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Display information
		t.Log(src, "=>", dest)
	}

}

func Test_rescaler_002(t *testing.T) {
	assert := assert.New(t)

	// Create an image generator
	image, err := generator.NewYUV420P(ffmpeg.VideoPar("yuv420p", "1280x720", 25))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer image.Close()

	// Create a rescaler
	rescaler, err := ffmpeg.NewRescaler(ffmpeg.VideoPar("rgb24", "1024x768", 25), false)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Temp output
	tmpdir, err := os.MkdirTemp("", t.Name())
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Rescale ten frames
	for i := 0; i < 10; i++ {
		f := image.Frame()
		if !assert.NotNil(f) {
			t.FailNow()
		}
		src_image, err := f.Image()
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Output as PNG
		tmpfile := filepath.Join(tmpdir, fmt.Sprintf("src_image_%03d", i)+".png")
		fsrc, err := os.Create(tmpfile)
		if !assert.NoError(err) {
			t.SkipNow()
		}
		defer fsrc.Close()
		err = png.Encode(fsrc, src_image)
		if !assert.NoError(err) {
			t.FailNow()
		}
		t.Logf("Wrote %s", tmpfile)

		// Rescale the frame
		dest, err := rescaler.Frame(f)
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Make a naive image
		dest_image, err := dest.Image()
		if !assert.NoError(err) {
			t.FailNow()
		}

		// Output as PNG
		tmpfile = filepath.Join(tmpdir, fmt.Sprintf("dest_image_%03d", i)+".png")
		fh, err := os.Create(tmpfile)
		if !assert.NoError(err) {
			t.SkipNow()
		}
		defer fh.Close()
		err = png.Encode(fh, dest_image)
		if !assert.NoError(err) {
			t.FailNow()
		}
		t.Logf("Wrote %s", tmpfile)
	}
}
