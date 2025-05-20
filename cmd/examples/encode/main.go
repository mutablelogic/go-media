package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
)

// This example encodes an audio an video stream to a file
func main() {
	// Check we have a filename
	if len(os.Args) != 2 {
		log.Fatal("Usage: encode filename")
	}

	// Create a new file with an audio and video stream
	file, err := ffmpeg.Create(os.Args[1],
		ffmpeg.OptStream(1, ffmpeg.VideoPar("yuv420p", "1280x720", 25, ffmpeg.NewMetadata("crf", 2))),
		ffmpeg.OptStream(2, ffmpeg.AudioPar("fltp", "mono", 22050)),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Make an video generator which can generate frames with the same
	// parameters as the video stream
	video, err := generator.NewEBU(file.Stream(1).Par())
	if err != nil {
		log.Fatal(err)
	}
	defer video.Close()

	// Make an audio generator which can generate a 1KHz tone
	// at -5dB with the same parameters as the audio stream
	audio, err := generator.NewSine(1000, -5, file.Stream(2).Par())
	if err != nil {
		log.Fatal(err)
	}
	defer audio.Close()

	// Bail out when we receive a signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGQUIT)
	defer cancel()

	// Write 90 seconds, passing video and audio frames to the encoder
	// and returning io.EOF when the duration is reached
	duration := float64(90)
	err = file.Encode(ctx, func(stream int) (*ffmpeg.Frame, error) {
		var frame *ffmpeg.Frame
		switch stream {
		case 1:
			frame = video.Frame()
		case 2:
			frame = audio.Frame()
		}
		if frame != nil && frame.Ts() < duration {
			fmt.Println(stream, frame.Ts())
			return frame, nil
		}
		return nil, io.EOF
	}, nil)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
	fmt.Print("\n")
}
