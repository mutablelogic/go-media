package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	// Packages
	"github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	file "github.com/mutablelogic/go-media/pkg/file"
	server "github.com/mutablelogic/go-server"
)

// Packages

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataCommands struct {
	Meta ListMetadata `cmd:"" group:"METADATA" help:"Examine metadata"`
}

type ListMetadata struct {
	Path      string `arg:"" group:"METADATA" type:"path" help:"Examine file metadata"`
	Recursive bool   `short:"r" help:"Recursively examine files"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListMetadata) Run(app server.Cmd) error {
	// Create the media manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(false, nil))
	if err != nil {
		return err
	}

	// Create a new file walker
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info os.FileInfo) error {
		if info.IsDir() {
			if !cmd.Recursive && relpath != "." {
				return file.SkipDir
			}
			return nil
		}

		// Open file
		f, err := manager.Open(filepath.Join(root, relpath), nil)
		if err != nil {
			return fmt.Errorf("%s: %w", info.Name(), err)
		}
		defer f.Close()

		// Print metadata
		result := make([]media.Metadata, 0, 20)
		result = append(result, ffmpeg.NewMetadata("path", filepath.Join(root, relpath)))

		if duration := f.(*ffmpeg.Reader).Duration(); duration > 0 {
			result = append(result, ffmpeg.NewMetadata("duration", duration.String()))
		}

		for _, meta := range f.(*ffmpeg.Reader).Metadata() {
			result = append(result, meta)
		}

		return write(os.Stdout, result, nil)
	})

	// Perform the walk, return any errors
	return walker.Walk(app.Context(), cmd.Path)
}
