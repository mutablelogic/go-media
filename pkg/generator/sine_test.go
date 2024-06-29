package generator_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mutablelogic/go-media/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func Test_sine_001(t *testing.T) {
	assert := assert.New(t)
	sine, err := generator.NewSine(2000, 10, 44100)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer sine.Close()

	t.Log(sine)
}

func Test_sine_002(t *testing.T) {
	assert := assert.New(t)
	sine, err := generator.NewSine(2000, 10, 44100)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer sine.Close()

	for i := 0; i < 10; i++ {
		frame := sine.Frame()
		t.Log(frame)
	}
}

func Test_sine_003(t *testing.T) {
	assert := assert.New(t)

	const sampleRate = 10000
	const frequency = 440
	const volume = -10.0

	sine, err := generator.NewSine(frequency, volume, sampleRate)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer sine.Close()

	tmpdir, err := os.MkdirTemp("", t.Name())
	if !assert.NoError(err) {
		t.SkipNow()
	}
	tmpfile := filepath.Join(tmpdir, "sine.f32le")
	fh, err := os.Create(tmpfile)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer fh.Close()

	var bytes_written int
	for {
		frame := sine.Frame()
		if frame.Time() > 10*time.Second {
			break
		}
		n, err := fh.Write(frame.Bytes(0))
		assert.NoError(err)
		bytes_written += n
	}

	t.Log("Wrote", bytes_written, " bytes to", tmpfile)
	t.Log("  play with: ffplay -f f32le -ar", sampleRate, "-ac 1", tmpfile)
}
