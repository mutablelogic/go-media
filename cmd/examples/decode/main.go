package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func main() {
	// Open a media file for reading. The format of the file is guessed.
	input, err := ffmpeg.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Make a map function which can be used to decode the streams and set
	// the parameters we want each audio and video stream to have.
	// The audio and video streams are resampled and resized to fit the
	// parameters we pass to the decoder.
	mapfunc := func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == VIDEO {
			// Convert frame to yuv420p as needed
			return ffmpeg.VideoPar("yuv420p", par.WidthHeight(), par.FrameRate()), nil
		}
		// Ignore other streams
		return nil, nil
	}

	// Make a folder where we're going to store the thumbnails
	tmp, err := os.MkdirTemp("", "decode")
	if err != nil {
		log.Fatal(err)
	}

	// Decode the streams and receive the video frame
	// If the map function is nil, the frames are copied. In this example,
	// we get a yuv420p frame at the same size as the original.
	n := 0
	err = input.Decode(context.Background(), mapfunc, func(stream int, frame *ffmpeg.Frame) error {
		// Write the frame to a file
		w, err := os.Create(filepath.Join(tmp, fmt.Sprintf("frame-%d.jpg", n)))
		if err != nil {
			return err
		}
		defer w.Close()

		// Coovert to an image and encode a JPEG
		if image, err := frame.Image(); err != nil {
			return err
		} else if err := jpeg.Encode(w, image, nil); err != nil {
			return err
		} else {
			log.Println("Wrote:", w.Name())
		}

		// End after 10 frames
		n++
		if n >= 10 {
			return io.EOF
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
