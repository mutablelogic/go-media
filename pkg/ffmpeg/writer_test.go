package ffmpeg_test

import (
	"fmt"
	"os"
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_writer_001(t *testing.T) {
	assert := assert.New(t)

	// Write to a file
	w, err := os.CreateTemp("", t.Name()+"_*.mp3")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer w.Close()

	// Create a writer with an audio stream
	writer, err := ffmpeg.NewWriter(w,
		ffmpeg.OptOutputFormat(w.Name()),
		ffmpeg.OptAudioStream(),
		ffmpeg.OptVideoStream("1280x720"),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	fmt.Println("Written to", w.Name())
}
