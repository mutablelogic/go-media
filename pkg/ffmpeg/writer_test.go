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
		ffmpeg.OptMetadata(ffmpeg.NewMetadata("title", t.Name())),
		ffmpeg.OptStream(1, ffmpeg.AudioPar("fltp", "mono", 22050)),
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

	// Write 15 mins of frames
	duration := 60 * time.Minute
	assert.NoError(writer.Encode(func(stream int) (*ff.AVFrame, error) {
		frame := audio.Frame()
		if frame.Time() >= duration {
			return nil, io.EOF
		} else {
			t.Log("Frame s", frame.Time().Truncate(time.Millisecond))
			return frame.(*ffmpeg.Frame).AVFrame(), nil
		}
	}, func(packet *ff.AVPacket, timebase *ff.AVRational) error {
		if packet != nil {
			t.Log("Packet", packet)
		}
		return writer.Write(packet)
	}))
	t.Log("Written to", w.Name())
}

func Test_writer_002(t *testing.T) {
	assert := assert.New(t)

	// Write to a file
	w, err := os.CreateTemp("", t.Name()+"_*.mp3")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer w.Close()

	// Create a writer with an audio stream
	writer, err := ffmpeg.Create(w.Name(),
		ffmpeg.OptMetadata(ffmpeg.NewMetadata("title", t.Name())),
		ffmpeg.OptStream(1, ffmpeg.AudioPar("fltp", "mono", 22050)),
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

	// Write 15 mins of frames
	duration := 15 * time.Minute
	assert.NoError(writer.Encode(func(stream int) (*ff.AVFrame, error) {
		frame := audio.Frame()
		if frame.Time() >= duration {
			return nil, io.EOF
		} else {
			t.Log("Frame s", frame.Time().Truncate(time.Millisecond))
			return frame.(*ffmpeg.Frame).AVFrame(), nil
		}
	}, func(packet *ff.AVPacket, timebase *ff.AVRational) error {
		if packet != nil {
			t.Log("Packet", packet)
		}
		return writer.Write(packet)
	}))
	t.Log("Written to", w.Name())
}

func Test_writer_003(t *testing.T) {
	assert := assert.New(t)

	// Write to a file
	w, err := os.CreateTemp("", t.Name()+"_*.ts")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer w.Close()

	// Create a writer with an audio stream
	writer, err := ffmpeg.NewWriter(w,
		ffmpeg.OptMetadata(ffmpeg.NewMetadata("title", t.Name())),
		ffmpeg.OptStream(1, ffmpeg.VideoPar("yuv420p", "1280x720", 25)),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Make an video generator
	video, err := generator.NewYUV420P(25, writer.Stream(1).Par())
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer video.Close()

	// Write 1 min of frames
	duration := time.Minute
	assert.NoError(writer.Encode(func(stream int) (*ff.AVFrame, error) {
		frame := video.Frame()
		if frame.Time() >= duration {
			return nil, io.EOF
		} else {
			t.Log("Frame", frame.Time().Truncate(time.Millisecond))
			return frame.(*ffmpeg.Frame).AVFrame(), nil
		}
	}, func(packet *ff.AVPacket, timebase *ff.AVRational) error {
		if packet != nil {
			t.Log("Packet", packet)
		}
		return writer.Write(packet)
	}))
	t.Log("Written to", w.Name())
}
