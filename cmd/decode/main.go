package main

import (
	"flag"
	"log"
	"os"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media"
)

var (
	in           = flag.String("in", "", "input file to decode")
	audio_stream = flag.Int("audio", -1, "audio stream to decode")
	video_stream = flag.Int("video", -1, "video stream to decode")
)

func main() {
	flag.Parse()

	// Check input file - read it
	if *in == "" {
		log.Fatal("-in flag must be specified")
	}
	r, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	input, err := ffmpeg.NewReader(r, "")
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()

	// Create a decoder for audio
	audio, err := input.NewDecoder(ffmpeg.AUDIO, *audio_stream)
	if err != nil {
		log.Fatal(err)
	} else if err := audio.ResampleS16Mono(22000); err != nil {
		log.Fatal(err)
	}

	// Create a decoder for video
	video, err := input.NewDecoder(ffmpeg.VIDEO, *video_stream)
	if err != nil {
		log.Fatal(err)
	} else if err := video.Rescale(1024, 720); err != nil {
		log.Fatal(err)
	}

	// Demux and decode the audio and video
	n := 0
	if err := input.Demux(input.Decode(func(frame ffmpeg.Frame) error {
		log.Print("frame: ", n, "=>", frame)
		n++
		return nil
	})); err != nil {

		log.Fatal(err)
	}
}
