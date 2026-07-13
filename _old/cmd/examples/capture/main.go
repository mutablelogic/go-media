package main

import (
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"regexp"
	"syscall"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

var (
	reDeviceNamePath = regexp.MustCompile(`^([a-z][a-zA-Z0-9]+)\:(.*)$`)
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: capture device:path")
	}

	// Get the format associated with the input file
	device := reDeviceNamePath.FindStringSubmatch(os.Args[1])
	if device == nil {
		log.Fatal("Invalid device name, use device:path")
	}

	// Create a media manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(false, nil))
	if err != nil {
		log.Fatal(err)
	}

	// Find device
	devices := manager.Formats(media.DEVICE, device[1])
	if len(devices) == 0 {
		log.Fatalf("No devices found for %v", device[1])
	}
	if len(devices) > 1 {
		log.Fatalf("Multiple devices found: %q", devices)
	}

	// Open device
	media, err := manager.Open(device[2], devices[0])
	if err != nil {
		log.Fatal(err)
	}
	defer media.Close()

	// Tmpdir
	tmpdir, err := os.MkdirTemp("", "capture")
	if err != nil {
		log.Fatal(err)
	}

	// Frame function
	frameFunc := func(stream int, frame *ffmpeg.Frame) error {
		w, err := os.Create(fmt.Sprintf("%v/frame-%v.jpg", tmpdir, frame.Ts()))
		if err != nil {
			return err
		}
		defer w.Close()

		image, err := frame.Image()
		if err != nil {
			return err
		}

		if err := jpeg.Encode(w, image, nil); err != nil {
			return err
		}

		fmt.Println("Written", w.Name())

		return nil
	}

	// Map function
	mapFunc := func(_ int, in *ffmpeg.Par) (*ffmpeg.Par, error) {
		fmt.Println("Input", in)
		return ffmpeg.VideoPar("yuv420p", in.WidthHeight(), in.FrameRate()), nil
	}

	// Receive frames
	if err := media.(*ffmpeg.Reader).Decode(
		ContextForSignal(os.Interrupt, syscall.SIGQUIT),
		mapFunc,
		frameFunc,
	); err != nil {
		log.Fatal(err)
	}
}
