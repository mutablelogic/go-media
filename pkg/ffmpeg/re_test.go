package ffmpeg_test

import (
	"context"
	"image/png"
	"os"
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_re_001(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../etc/test/sample.png")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()
	img, err := png.Decode(r)
	if !assert.NoError(err) {
		t.FailNow()
	}
	frame, err := ffmpeg.FrameFromImage(img)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Create a resizer
	re, err := ffmpeg.NewRe(ffmpeg.VideoPar("gray8", "vga", 25), false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer re.Close()

	// Resize the frame
	dest, err := re.Frame(frame)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Get image from frame
	destimg, err := dest.ImageFromFrame()
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Save image
	w, err := os.CreateTemp("", "*_resized.png")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer w.Close()

	err = png.Encode(w, destimg)
	if !assert.NoError(err) {
		t.FailNow()
	}

	t.Log(r.Name(), "=>", w.Name())
}

func Test_re_002(t *testing.T) {
	assert := assert.New(t)

	r, err := ffmpeg.Open("../../etc/test/sample.mp3")
	//r, err := ffmpeg.Open("/Volumes/Drobo/Media/Music/ABBA/Gold_ Greatest Hits/01 Dancing Queen.m4a")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Make resampler
	re, err := ffmpeg.NewRe(ffmpeg.AudioPar("s16", "stereo", 44100), true)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer re.Close()

	// Write out resampled audio
	w, err := os.CreateTemp("", "*_resampled.sw")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer w.Close()

	// Decode function
	decodefn := func(_ int, frame *ffmpeg.Frame) error {
		if frame != nil && frame.Type() != ffmpeg.AUDIO {
			return nil
		}
		resampled, err := re.Frame(frame)
		if err != nil {
			return err
		}
		if resampled != nil {
			if _, err := w.Write(resampled.Bytes(0)); err != nil {
				return err
			}
		}
		return nil
	}
	// Get audio frames
	if err := r.Decode(context.Background(), decodefn, nil); !assert.NoError(err) {
		t.FailNow()
	}
	// Print
	t.Log("  play with: ffplay -f s16le -ar 44100 -ac 2", w.Name())

}
