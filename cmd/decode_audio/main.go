package main

import (
	"flag"
	"log"
	"os"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

var (
	in     = flag.String("in", "", "input file to decode")
	stream = flag.Int("stream", -1, "stream to decode")
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
	_, err = input.NewDecoder(ffmpeg.AUDIO, *stream)
	if err != nil {
		log.Fatal(err)
	}

	// Demux and decode
	if err := input.Demux(input.Decode(func(frame ffmpeg.Frame) error {
		log.Printf("frame: %v", frame)
		return nil
	})); err != nil {
		log.Fatal(err)
	}
}
