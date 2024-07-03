package generator_test

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	"github.com/mutablelogic/go-media/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func Test_ebu_001(t *testing.T) {
	assert := assert.New(t)
	image, err := generator.NewEBU(ffmpeg.VideoPar("rgb24", "1280x720", 25))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer image.Close()

	t.Log(image)
}

func Test_ebu_002(t *testing.T) {
	assert := assert.New(t)
	image, err := generator.NewEBU(ffmpeg.VideoPar("rgb24", "1280x720", 25))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer image.Close()

	tmpdir, err := os.MkdirTemp("", t.Name())
	if !assert.NoError(err) {
		t.SkipNow()
	}

	for i := 0; i < 30; i++ {
		img, err := image.Frame().Image()
		if !assert.NoError(err) {
			t.FailNow()
		}
		tmpfile := filepath.Join(tmpdir, fmt.Sprintf("image_%03d", i)+".png")
		fh, err := os.Create(tmpfile)
		if !assert.NoError(err) {
			t.SkipNow()
		}
		defer fh.Close()
		err = png.Encode(fh, img)
		if !assert.NoError(err) {
			t.FailNow()
		}
		t.Logf("Wrote %s", tmpfile)
	}
}
