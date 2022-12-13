/*
An example of extracting frames from a video. Provide a single video file on the command line,
and it will extract and save individual frames to the current working directory, or the directory
specified by the -out flag.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	// Packages
	config "github.com/mutablelogic/go-media/pkg/config"
	media "github.com/mutablelogic/go-media/pkg/media"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

var (
	flagVersion  = flag.Bool("version", false, "Print version information")
	flagDebug    = flag.Bool("debug", false, "Enable debug output")
	flagOut      = flag.String("out", "", "Output filename")
	flagAudio    = flag.Bool("audio", false, "Extract audio")
	flagVideo    = flag.Bool("video", false, "Extract video")
	flagSubtitle = flag.Bool("subtitle", false, "Extract subtitles")
)

func main() {
	flag.Parse()

	// Check for version
	if *flagVersion {
		config.PrintVersion(flag.CommandLine.Output())
		media.PrintVersion(flag.CommandLine.Output())
		os.Exit(0)
	}

	// Check output arguments
	if flag.NArg() != 1 || *flagOut == "" {
		flag.Usage()
		os.Exit(-1)
	}

	// Create a cancellable context
	ctx := ContextForSignal(os.Interrupt)

	// Create a media manager, set debugging
	manager := media.New()
	manager.SetDebug(*flagDebug)

	// Open the output file
	out, err := manager.CreateFile(*flagOut)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
	fmt.Println(out)

	// Open the input file
	media, err := manager.OpenFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// Create a media map
	flags := MEDIA_FLAG_NONE
	if *flagAudio {
		flags |= MEDIA_FLAG_AUDIO
	}
	if *flagVideo {
		flags |= MEDIA_FLAG_VIDEO
	}
	if *flagSubtitle {
		flags |= MEDIA_FLAG_SUBTITLE
	}
	media_map, err := manager.Map(media, flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// Demux the media
	if err := manager.Demux(ctx, media_map, func(_ context.Context, packet Packet) error {
		fmt.Println("Packet", packet)
		return manager.Decode(ctx, media_map, packet, func(_ context.Context, frame Frame) error {
			fmt.Println("  Frame", frame)
			return nil
		})
	}); err != nil && err != context.Canceled {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// Close the output file
	if err := out.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// If the context was cancelled, print a message
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "\nInterrupted")
	}
}

// ContextForSignal returns a context object which is cancelled when a signal
// is received. It returns nil if no signal parameter is provided
func ContextForSignal(signals ...os.Signal) context.Context {
	if len(signals) == 0 {
		return nil
	}

	ch := make(chan os.Signal)
	ctx, cancel := context.WithCancel(context.Background())

	// Send message on channel when signal received
	signal.Notify(ch, signals...)

	// When any signal received, call cancel
	go func() {
		<-ch
		cancel()
	}()

	// Return success
	return ctx
}
