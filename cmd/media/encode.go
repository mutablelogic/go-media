package main

import (
	"io"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type EncodeCommands struct {
	EncodeTest EncodeTest `cmd:"" group:"TRANSCODE" help:"Encode a test file"`
}

type EncodeTest struct {
	Out      string        `arg:"" type:"path" help:"Output filename"`
	Duration time.Duration `help:"Duration of the test file"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *EncodeTest) Run(app server.Cmd) error {
	// Create a manager
	manager, err := ffmpeg.NewManager()
	if err != nil {
		return err
	}

	// Streams
	streams := []media.Par{
		manager.MustVideoPar("yuv420p", 1280, 720, 25),
		manager.MustAudioPar("fltp", "mono", 22050),
	}

	// Create a writer with two streams
	writer, err := manager.Create(cmd.Out, nil, nil, streams...)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Make an video generator which can generate frames with the same parameters as the video stream
	video, err := generator.NewEBU(writer.(*ffmpeg.Writer).Stream(1).Par())
	if err != nil {
		return err
	}
	defer video.Close()

	// Make an audio generator which can generate a 1KHz tone
	// at -5dB with the same parameters as the audio stream
	audio, err := generator.NewSine(1000, -5, writer.(*ffmpeg.Writer).Stream(2).Par())
	if err != nil {
		return err
	}
	defer audio.Close()

	// Write until CTRL+C or duration is reached
	manager.Errorf("Press CTRL+C to stop encoding\n")
	var ts uint
	return manager.Encode(app.Context(), writer, func(stream int) (media.Frame, error) {
		var frame *ffmpeg.Frame
		switch stream {
		case 1:
			frame = video.Frame()
		case 2:
			frame = audio.Frame()
		}

		// Print the timestamp in seconds
		if newts := uint(frame.Ts()); newts != ts {
			ts = newts
			manager.Errorf("Writing frame at %s\r", time.Duration(ts)*time.Second)
		}

		// Check for end of stream
		if cmd.Duration == 0 || frame.Ts() < cmd.Duration.Seconds() {
			return frame, nil
		} else {
			return frame, io.EOF
		}
	})
}
