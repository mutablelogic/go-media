package main

import (
	"context"
	"errors"
	"fmt"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	// Packages
	"github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/file"
)

type ThumbnailsCmd struct {
	Path  string        `arg:"" required:"" help:"Media file, device or path" type:"string"`
	Dur   time.Duration `name:"duration" help:"Duration between thumnnails" type:"duration" default:"1m"`
	Width int           `name:"width" help:"Width of thumbnail" type:"int" default:"320"`
}

func (cmd *ThumbnailsCmd) Run(globals *Globals) error {
	// If we have a device, then use this
	format, path := formatFromPath(globals.manager, media.NONE, cmd.Path)
	if format != nil {
		return cmd.mediaWalker(globals.ctx, globals.manager, format, path)
	}

	// Create the walker with the processor callback
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info fs.FileInfo) error {
		if info.IsDir() || info.Size() == 0 {
			return nil
		}
		if err := cmd.mediaWalker(ctx, globals.manager, nil, filepath.Join(root, relpath)); err != nil {
			if err == context.Canceled {
				globals.manager.Infof("Cancelled\n")
			} else {
				globals.manager.Errorf("Error processing %q: %v\n", relpath, err)
			}
		}
		return nil
	})

	// Walk the filesystem
	return walker.Walk(globals.ctx, cmd.Path)
}

func (cmd *ThumbnailsCmd) mediaWalker(ctx context.Context, manager media.Manager, format media.Format, path string) error {
	reader, err := manager.Open(path, format)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create a decoder for video - output 320x240 frames
	// We should adjust this to match the output size of the thumbnail
	decoder, err := reader.Decoder(func(stream media.Stream) (media.Parameters, error) {
		if stream.Type().Is(media.VIDEO) {
			// TODO: We need to use the sample aspect ratio
			width := cmd.Width
			height := stream.Parameters().Height() * width / stream.Parameters().Width()
			return manager.VideoParameters(width, height, "rgb24")
		} else {
			return nil, nil
		}
	})
	if err != nil {
		return err
	}

	// Decode the frames
	var t time.Duration = -1
	return decoder.Decode(ctx, func(frame media.Frame) error {
		// Logic to return if we have a frame within the duration
		if frame.Time() < 0 {
			return nil
		} else if t != -1 && frame.Time()-t < cmd.Dur {
			return nil
		} else {
			t = frame.Time()
		}

		// Save the frame
		filename := fmt.Sprintf("%s.%s.png", filepath.Base(path), t.Truncate(time.Second))
		image, err := frame.Image()
		if err != nil {
			return err
		}

		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()

		fmt.Println("Writing", filename)
		if err := png.Encode(w, image); err != nil {
			return errors.Join(err, os.Remove(filename))
		}

		return nil
	})
}
