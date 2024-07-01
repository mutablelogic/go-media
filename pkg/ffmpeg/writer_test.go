package ffmpeg_test

import (
	"io"
	"os"
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
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
	duration := float64(15 * 60)
	assert.NoError(writer.Encode(func(stream int) (*ffmpeg.Frame, error) {
		frame := audio.Frame()
		if frame.Ts() >= duration {
			return nil, io.EOF
		} else {
			t.Log("Frame", frame.Ts())
			return frame, nil
		}
	}, func(packet *ffmpeg.Packet) error {
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
	duration := float64(15 * 60)
	assert.NoError(writer.Encode(func(stream int) (*ffmpeg.Frame, error) {
		frame := audio.Frame()
		if frame.Ts() >= duration {
			return nil, io.EOF
		} else {
			t.Log("Frame ", frame.Ts())
			return frame, nil
		}
	}, func(packet *ffmpeg.Packet) error {
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
		ffmpeg.OptStream(1, ffmpeg.VideoPar("yuv420p", "640x480", 30)),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Make an video generator
	video, err := generator.NewYUV420P(writer.Stream(1).Par())
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer video.Close()

	// Write 1 min of frames
	duration := float64(60)
	assert.NoError(writer.Encode(func(stream int) (*ffmpeg.Frame, error) {
		frame := video.Frame()
		if frame.Ts() >= duration {
			return nil, io.EOF
		} else {
			t.Log("Frame", stream, "=>", frame.Ts())
			return frame, nil
		}
	}, func(packet *ffmpeg.Packet) error {
		if packet != nil {
			t.Log("Packet", packet.Ts())
		}
		return writer.Write(packet)
	}))
	t.Log("Written to", w.Name())
}

func Test_writer_004(t *testing.T) {
	assert := assert.New(t)

	// Write to a file
	w, err := os.CreateTemp("", t.Name()+"_*.m4v")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer w.Close()

	// Create a writer with an audio stream
	writer, err := ffmpeg.Create(w.Name(),
		ffmpeg.OptMetadata(ffmpeg.NewMetadata("title", t.Name())),
		ffmpeg.OptStream(1, ffmpeg.VideoPar("yuv420p", "640x480", 30)),
		ffmpeg.OptStream(2, ffmpeg.AudioPar("fltp", "mono", 22050)),
	)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Make an video generator
	video, err := generator.NewYUV420P(writer.Stream(1).Par())
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer video.Close()

	// Make an audio generator
	audio, err := generator.NewSine(440, -5, writer.Stream(2).Par())
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Write 10 secs of frames
	duration := float64(10)
	assert.NoError(writer.Encode(func(stream int) (*ffmpeg.Frame, error) {
		var frame *ffmpeg.Frame
		switch stream {
		case 1:
			frame = video.Frame()
		case 2:
			frame = audio.Frame()
		}
		if frame.Ts() >= duration {
			t.Log("Frame time is EOF", frame.Ts())
			return nil, io.EOF
		} else {
			t.Log("Frame", stream, "=>", frame.Ts())
			return frame, nil
		}
	}, nil))
	t.Log("Written to", w.Name())
}
