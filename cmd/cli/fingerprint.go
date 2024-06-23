package main

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	// Packages
	"github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/chromaprint"
	"github.com/mutablelogic/go-media/pkg/file"
)

type FimgerprintCmd struct {
	Path string `arg:"" required:"" help:"Media file or path" type:"path"`
}

func (cmd *FimgerprintCmd) Run(globals *Globals) error {
	// Create the walker with the processor callback
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info fs.FileInfo) error {
		if info.IsDir() || info.Size() == 0 {
			return nil
		}
		if err := cmd.mediaWalker(ctx, globals.manager, filepath.Join(root, relpath)); err != nil {
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

func (cmd *FimgerprintCmd) mediaWalker(ctx context.Context, manager media.Manager, path string) error {
	reader, err := manager.Open(path, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create a decoder for audio - needs to be s16 pcm for chromaprint
	decoder, err := reader.Decoder(func(stream media.Stream) (media.Parameters, error) {
		if stream.Type().Is(media.AUDIO) {
			return manager.AudioParameters("mono", "s16", 22050)
		} else {
			return nil, nil
		}
	})
	if err != nil {
		return err
	}

	// Create a fingerprinter
	fingerprint := chromaprint.New(22050, 1, time.Minute)
	defer fingerprint.Close()

	// Decode the frames
	if err := decoder.Decode(ctx, func(frame media.Frame) error {
		_, err := fingerprint.Write(frame.Int16(0))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if hash, err := fingerprint.Finish(); err != nil {
		return err
	} else {
		fmt.Println(path, "=>", hash)
	}

	// Return success
	return nil
}
