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
	"regexp"

	// Packages
	config "github.com/mutablelogic/go-media/pkg/config"
	media "github.com/mutablelogic/go-media/pkg/media"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

var (
	reDeviceName = regexp.MustCompile(`^([A-Za-z]\w+)$`)
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

	// The output is a path, URL, or the name of an output device
	if flags.Out() == "" {
		fmt.Fprintln(os.Stderr, "No output specified")
		os.Exit(-2)
	}

	// Check to force a specific input format
	var in MediaFormat
	var media Media
	if flags.In() != "" {
		if formats := manager.MediaFormats(MEDIA_FLAG_DECODER, flags.In()); len(formats) == 0 {
			fmt.Fprintf(os.Stderr, "No input format found for %q", flags.In())
			os.Exit(-2)
		} else if len(formats) > 1 {
			fmt.Fprintf(os.Stderr, "Multiple input formats found for %q", flags.In())
			os.Exit(-1)
		} else {
			in = formats[0]
		}
	}

	// Check for a device, URL or filepath
	if reDeviceName.MatchString(flags.Arg(0)) {
		media, err = manager.OpenDevice(flags.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-2)
		}
	} else if url, err := url.Parse(flags.Arg(0)); err == nil && url.Scheme != "" {
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

	// Create a map of the streams for the input file
	media_map, err := manager.Map(media, flags.MediaFlags())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// For the output, we can output to a device, a URL, or a file
	var out Media
	if reDeviceName.MatchString(flags.Out()) {
		out, err = manager.CreateDevice(flags.Out())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-2)
		}
	} else if url, err := url.Parse(flags.Out()); err == nil && url.Scheme != "" {
		fmt.Fprintln(os.Stderr, "Output URL's are not yet supported")
		os.Exit(-2)
	} else if out, err = manager.CreateFile(flags.Out()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// For each input stream, add an output
	for _, stream := range media_map.Streams(flags.MediaFlags()) {
		// TODO
		// If audio, then resample data
		switch {
		case stream.Flags().Is(MEDIA_FLAG_AUDIO):
			if err := media_map.Resample(AudioFormat{
				Rate:   11025,
				Format: SAMPLE_FORMAT_U8,
			}, stream); err != nil {
				fmt.Fprintln(os.Stderr, "Cannot resample audio stream:", err)
				os.Exit(-2)
			}
		}
	}

	// Print the map
	media_map.PrintMap(os.Stdout)

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
	if err := out.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

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
