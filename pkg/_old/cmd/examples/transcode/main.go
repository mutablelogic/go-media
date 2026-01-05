package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

// This example encodes an audio an video stream to a file
func main() {
	// Bail out when we receive a signal
	ctx := ContextForSignal(os.Interrupt, syscall.SIGQUIT)

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

	// Map to output file
	decoding, err := in.Map(func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		// This is where you specify the output format for the input stream
		return par, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create an output file from from the map
	out, err := ffmpeg.Create(os.Args[2], ffmpeg.OptContext(decoding))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	// Decoding goroutine
	var result error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// This is where we decode the stream
		result = errors.Join(result, in.DecodeWithContext(ctx, decoding, func(stream int, frame *ffmpeg.Frame) error {

			// Add the frame onto the encoding queue
			fmt.Println("->DECODE:", stream, frame)
			decoding.C(stream) <- frame
			fmt.Println("<-DECODE")

			// Return success
			return nil
		}))
	}()

	// Encoding goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		// This is where we encode the stream
		result = errors.Join(result, out.Encode(ctx, func(stream int) (*ffmpeg.Frame, error) {
			fmt.Println("->ENCODE", stream)
			select {
			case frame := <-decoding.C(stream):
				fmt.Println("<-ENCODE", frame)
				return frame, nil
			default:
				// Not ready to pass a frame back
				fmt.Println("<-ENCODE", nil)
				return nil, nil
			}
		}, nil))
		if result != nil {
			fmt.Println("ERROR:", result)
		}
	}()

	wg.Wait()

	if result != nil {
		log.Fatal(result)
	}
	fmt.Println("Transcoded to", out)
}
