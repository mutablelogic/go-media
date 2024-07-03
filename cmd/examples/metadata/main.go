package main

import (
	"log"
	"os"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

func main() {
	// Open a media file for reading. The format of the file is guessed.
	reader, err := ffmpeg.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// Retrieve all the metadata from the file, and display it. If you pass
	// keys to the Metadata function, then only entries with those keys will be
	// returned.
	for _, metadata := range reader.Metadata() {
		log.Print(metadata.Key(), " => ", metadata.Value())
	}

	// Retrieve artwork by using the MetaArtwork key. The value is of type []byte.
	// which needs to be converted to an image. There is a utility method to
	// detect the image type.
	for _, artwork := range reader.Metadata(ffmpeg.MetaArtwork) {
		mimetype := artwork.Value()
		if mimetype != "" {
			log.Print("We got some artwork of mimetype ", mimetype)
		}
	}

}
