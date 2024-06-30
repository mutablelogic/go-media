package ffmpeg_test

import (
	"io"
	"os"
	"testing"
	"time"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
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
		ffmpeg.OptAudioStream(1, ffmpeg.AudioPar("fltp", "mono", 22050)),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Make an audio generator
	audio, err := generator.NewSine(440, -5, writer.Stream(1).Par())
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer audio.Close()

	// Write frames
	n := 0
	assert.NoError(writer.Encode(func(stream int) (*ff.AVFrame, error) {
		frame := audio.Frame()
		t.Log("Frame s", frame.Time().Truncate(time.Millisecond))
		if frame.Time() > 10*time.Second {
			return nil, io.EOF
		} else {
			return frame.(*ffmpeg.Frame).AVFrame(), nil
		}
	}, func(packet *ff.AVPacket) error {
		if packet != nil {
			t.Log("Packet ts", packet.Pts())
			n += packet.Size()
		}
		return writer.Write(packet)
	}))
	t.Log("Written", n, "bytes to", w.Name())
}
