/*
An example of extracting artwork from media files. Provide a list of files or directories to process
on the command line, and the artwork will be extracted to the same directory as the media file,
unless the -out flag is provided, in which case the artwork will be written to the specified directory.

The command attempts not to write the same file twice, so if you run it twice on the same directory,
you will only get new artwork files.
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	config "github.com/mutablelogic/go-media/pkg/config"
	file "github.com/mutablelogic/go-media/pkg/file"
	media "github.com/mutablelogic/go-media/pkg/media"
)

var (
	flagDebug   = flag.Bool("debug", false, "Enable debug output")
	flagOut     = flag.String("out", "", "Output directory for artwork")
	flagVersion = flag.Bool("version", false, "Print version information")
)

func main() {
	flag.Parse()

	// Check for version
	if *flagVersion {
		config.PrintVersion(flag.CommandLine.Output())
		media.PrintVersion(flag.CommandLine.Output())
		os.Exit(0)
	}

	// Check arguments
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	// Create a cancellable context
	ctx := ContextForSignal(os.Interrupt)

	// Create a media manager, set debugging
	manager := media.New()
	manager.SetDebug(*flagDebug)

	// Check the out directory
	if *flagOut != "" {
		if info, err := os.Stat(*flagOut); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-2)
		} else if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "%q: not a directory\n", info.Name())
			os.Exit(-2)
		}
	}

	// Create a file walker object
	var result error
	walker := file.NewWalker(func(ctx context.Context, root, path string, info fs.FileInfo) error {
		// Ignore folders
		if info.IsDir() {
			return nil
		}

		// open the file
		media, err := manager.OpenFile(filepath.Join(root, path))
		if err != nil {
			result = multierror.Append(result, err)
		}
		defer media.Close()

		// Folder we should write to
		out := *flagOut
		if *flagOut == "" {
			out = filepath.Dir(filepath.Join(root, path))
		}

		// Process the media file
		if err := ProcessMedia(ctx, out, media); err != nil {
			result = multierror.Append(result, err)
		}

		// Always return success
		return nil
	})

	// Process each path
	for _, path := range flag.Args() {
		if err := walker.Walk(ctx, path); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-3)
		}
	}

	// If the context was cancelled, print a message
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "\nInterrupted")
	}

	// Report errors
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
		os.Exit(-3)
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
