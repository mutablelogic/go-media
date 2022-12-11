/*
An example of extracting artwork from media files. Provide a list of files or directories to process
on the command line, and the artwork will be extracted to the same directory as the media file,
unless the -out flag is provided, in which case the artwork will be written to the specified directory.
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
	"github.com/hashicorp/go-multierror"
	"github.com/mutablelogic/go-media/pkg/file"
	"github.com/mutablelogic/go-media/pkg/media"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func main() {
	// Check arguments
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Create a cancellable context
	ctx := ContextForSignal(os.Interrupt)

	// Create a media manager
	manager := media.New()

	// Create a file walker object
	var result error
	walker := file.NewWalker(func(ctx context.Context, root, path string, info fs.FileInfo) error {
		if info.IsDir() {
			return nil
		}
		// open the file
		media, err := manager.OpenFile(filepath.Join(root, path))
		if err != nil {
			result = multierror.Append(result, err)
		}
		defer media.Close()

		// Process the media file
		if err := ProcessMedia(ctx, media); err != nil {
			result = multierror.Append(result, err)
		}

		// Always return success
		return nil
	})

	// Process each path
	count := 0
	for _, path := range flag.Args() {
		if err := walker.Walk(ctx, path); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		count += walker.Count()
	}

	// If the context was cancelled, print a message
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "\nInterrupted")
	}

	// Print number of items processed, and exit successfully
	fmt.Println(count, "items processed")
	os.Exit(0)
}

// ProcessMedia processes media files through the pipeline
func ProcessMedia(ctx context.Context, media Media) error {
	fmt.Println(media.URL())
	return nil
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
