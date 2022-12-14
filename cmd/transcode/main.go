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
	"net/url"
	"os"
	"os/signal"

	// Packages
	config "github.com/mutablelogic/go-media/pkg/config"
	media "github.com/mutablelogic/go-media/pkg/media"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func main() {
	// Create a media manager object
	manager := media.New()
	defer manager.Close()

	flags, err := NewFlags(os.Args)
	if err != nil {
		// Check for -help -version, etc.
		if err != flag.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		} else {
			os.Exit(0)
		}
	}

	// If there is a version flag, print version and exit
	if flags.Version() {
		config.PrintVersion(flags.Writer())
		media.PrintVersion(flags.Writer())
		os.Exit(0)
	}

	// If there are no arguments but there is a -in flag, then list input formats
	if flags.NArg() == 0 {
		count := 0
		if flags.In() != "" {
			count += enumerateFormats(flags, manager, MEDIA_FLAG_DECODER, flags.In(), "Inputs:")
		}
		if flags.Out() != "" {
			count += enumerateFormats(flags, manager, MEDIA_FLAG_ENCODER, flags.Out(), "Outputs:")
		}
		if count == 0 {
			flags.PrintShortUsage()
			os.Exit(-1)
		} else {
			os.Exit(0)
		}
	}

	// Set the debug flag
	manager.SetDebug(flags.Debug())

	// Check to force a specific input format
	var in MediaFormat
	var media Media
	if flags.In() != "" {
		if formats := manager.MediaFormats(MEDIA_FLAG_DECODER, flags.In()); len(formats) == 0 {
			fmt.Fprintf(os.Stderr, "No input format found for %q", flags.In())
			os.Exit(-1)
		} else if len(formats) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple input formats found for %q", flags.In())
			os.Exit(-1)
		} else {
			in = formats[0]
		}
	}

	// Check for a URL or filepath
	if url, err := url.Parse(flags.Arg(0)); err == nil && url.Scheme != "" {
		media, err = manager.OpenURL(flags.Arg(0), in)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-2)
		}
	} else {
		media, err = manager.OpenFile(flags.Arg(0), in)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-2)
		}
	}
	defer media.Close()

	media_map, err := manager.Map(media, flags.MediaFlags())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
	fmt.Println(media_map)

	// Create a cancellable context
	ctx := contextForSignal(os.Interrupt)

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
	//if err := out.Close(); err != nil {
	//	fmt.Fprintln(os.Stderr, err)
	//	os.Exit(-2)
	//}

	// If the context was cancelled, print a message
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "\nInterrupted")
	}

}

func enumerateFormats(flags *Flags, manager Manager, flag MediaFlag, name, prefix string) int {
	var args []string
	if name != "" && name != "*" {
		args = append(args, name)
	}
	formats := manager.MediaFormats(flag, args...)
	if len(formats) > 0 {
		flags.PrintFormats(prefix, formats)
	}
	return len(formats)
}

// contextForSignal returns a context object which is cancelled when a signal
// is received. It returns nil if no signal parameter is provided
func contextForSignal(signals ...os.Signal) context.Context {
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

/*

	// Open the output file
	out, err := manager.CreateFile(*flagOut)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// Set output timestamp
	out.Set(MEDIA_KEY_CREATED, time.Now())

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

*/
