package main

import (
	"io"
	"log"
	"os"
	"time"

	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	generator "github.com/mutablelogic/go-media/pkg/generator"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func main() {
	// Create a new file with an audio and video stream
	file, err := ffmpeg.Create(os.Args[1],
		ffmpeg.OptStream(1, ffmpeg.VideoPar("yuv420p", "640x480", 30)),
		ffmpeg.OptStream(2, ffmpeg.AudioPar("fltp", "mono", 22050)),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Make an video generator which can generate YUV420P frames
	// with the same parameters as the video stream
	video, err := generator.NewYUV420P(file.Stream(1).Par())
	if err != nil {
		log.Fatal(err)
	}
	defer video.Close()

	// Make an audio generator which can generate a 440Hz tone
	// at -5dB with the same parameters as the audio stream
	audio, err := generator.NewSine(440, -5, file.Stream(2).Par())
	if err != nil {
		log.Fatal(err)
	}
	defer audio.Close()

	// Write 1 min of frames, passing video and audio frames to the encoder
	// and returning io.EOF when the duration is reached
	duration := time.Minute
	if err := file.Encode(func(stream int) (*ff.AVFrame, error) {
		var frame media.Frame
		switch stream {
		case 1:
			frame = video.Frame()
		case 2:
			frame = audio.Frame()
		}
		if frame.Time() >= duration {
			return nil, io.EOF
		} else {
			log.Println("Frame", stream, "=>", frame.Time().Truncate(time.Millisecond))
			return frame.(*ffmpeg.Frame).AVFrame(), nil
		}
	}, nil); err != nil {
		log.Fatal(err)
	}
}
