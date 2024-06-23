package main

import (
	"log"
	"os"

	media "github.com/mutablelogic/go-media"
	file "github.com/mutablelogic/go-media/pkg/file"
)

func main() {
	manager, err := media.NewManager()
	if err != nil {
		log.Fatal(err)
	}

	// Open a media file for reading. The format of the file is guessed.
	// Alteratively, you can pass a format as the second argument. Further optional
	// arguments can be used to set the format options.
	reader, err := manager.Open(os.Args[1], nil)
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
	for _, artwork := range reader.Metadata(media.MetaArtwork) {
		mimetype, ext, err := file.MimeType(artwork.Value().([]byte))
		if err != nil {
			log.Fatal(err)
		}
		log.Print("got artwork", mimetype, ext)
	}

}
