package main

import (
	"context"
	"log"
	"os"

	media "github.com/mutablelogic/go-media"
)

func main() {
	manager, err := media.NewManager()
	if err != nil {
		log.Fatal(err)
	}

	// Open a media file for reading. The format of the file is guessed.
	// Alteratively, you can pass a format as the second argument. Further optional
	// arguments can be used to set the format options.
	file, err := manager.Open(os.Args[1], nil)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Choose which streams to demultiplex - pass the stream parameters
	// to the decoder. If you don't want to resample or reformat the streams,
	// then you can pass nil as the function and all streams will be demultiplexed.
	decoder, err := file.Decoder(func(stream media.Stream) (media.Parameters, error) {
		// Copy streams, don't resample or resize
		return stream.Parameters(), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Demuliplex the stream and receive the frames of audio and video.
	if err := decoder.Decode(context.Background(), func(frame media.Frame) error {
		// Each packet is specific to a stream. It can be processed here
		// to receive audio or video frames, then resize or resample them,
		// for example. Alternatively, you can pass the packet to an encoder
		// to remultiplex the streams without processing them.
		log.Print(frame)

		// Return io.EOF to stop processing, nil to continue
		return nil
	}); err != nil {
		log.Fatal(err)
	}
}
