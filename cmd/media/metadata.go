package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Packages
	"github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	file "github.com/mutablelogic/go-media/pkg/file"
	server "github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/types"
)

// Packages

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataCommands struct {
	Meta    ListMetadata `cmd:"" group:"METADATA" help:"Examine metadata"`
	Artwork ListArtwork  `cmd:"" group:"METADATA" help:"Extract artwork"`
}

type ListMetadata struct {
	Path      string `arg:"" type:"path" help:"File or directory"`
	Recursive bool   `short:"r" help:"Recursively examine files"`
}

type ListArtwork struct {
	Path      string `arg:"" type:"path" help:"File or directory"`
	Recursive bool   `short:"r" help:"Recursively examine files"`
	Out       string `required:"" help:"Output filename for artwork, relative to the source path. Use {count} {hash} {path} {name} or {ext} for placeholders" default:"{hash}{ext}"`
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
		result = append(result, ffmpeg.NewMetadata("type", f.Type()))

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

func (cmd *ListArtwork) Run(app server.Cmd) error {
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

		// Extract artwork
		count := 1
		result := make([]media.Metadata, 0, 20)
		for _, meta := range f.(*ffmpeg.Reader).Metadata("artwork") {
			data := meta.Bytes()
			mimetype, ext, err := file.MimeType(data)
			if err != nil {
				return err
			}

			// Output the file
			out := template(cmd.Out, "hash", types.Hash(data), "path", filepath.Dir(relpath), "name", info.Name(), "mimetype", mimetype, "ext", ext, "count", count)

			// If the filename is relative, make it absolute
			if !filepath.IsAbs(out) {
				out = filepath.Join(root, out)
			}

			// If file exists, skip it
			if stat, err := os.Stat(out); err == nil && stat.Mode().IsRegular() && stat.Size() == int64(len(data)) {
				continue
			}

			// Make the directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Writing %s\n", out)
			if err := os.WriteFile(out, data, 0644); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			count++
		}

		return write(os.Stdout, result, nil)
	})

	// Perform the walk, return any errors
	return walker.Walk(app.Context(), cmd.Path)
}

func template(tmpl string, args ...any) string {
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			tmpl = strings.ReplaceAll(tmpl, fmt.Sprintf("{%s}", args[i]), fmt.Sprint(args[i+1]))
		}
	}
	return tmpl
}
