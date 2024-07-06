package main

import (
	"log"
	"os"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

// This example encodes an audio an video stream to a file
func main() {
	// Check we have a filename
	if len(os.Args) != 3 {
		log.Fatal("Usage: transcode <in> <out>")
	}

	// Read the input
	in, err := ffmpeg.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()
}
