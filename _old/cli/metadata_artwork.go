package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	// Packages

	"github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/file"
)

type MetadataCmd struct {
	Path string `arg:"" required:"" help:"Media file or directory of files" type:"path"`
}

type ArtworkCmd struct {
	Path string `arg:"" required:"" help:"Media file" type:"path"`
}

func (cmd *MetadataCmd) Run(globals *Globals) error {
	// Create the walker with the processor callback
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info fs.FileInfo) error {
		// Ignore directories
		if info.IsDir() {
			return nil
		}

		// Open the file
		reader, err := globals.manager.Open(filepath.Join(root, relpath), nil)
		if err != nil {
			globals.manager.Errorf("Error opening %q: %v\n", relpath, err)
			return nil
		}
		defer reader.Close()

		fmt.Println(info.Name())
		for _, entry := range reader.Metadata() {
			fmt.Println(entry)
		}

		return nil
	})

	if err := walker.Walk(globals.ctx, cmd.Path); err != nil {
		return err
	}

	return nil
}

func (cmd *ArtworkCmd) Run(globals *Globals) error {
	manager := globals.manager
	reader, err := manager.Open(cmd.Path, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	artworks := reader.Metadata(media.MetaArtwork)
	if len(artworks) == 0 {
		return errors.New("no artwork")
	}

	for i, artwork := range artworks {
		data := artwork.Value().([]byte)
		_, ext, err := file.MimeType(data)
		if err != nil {
			return err
		} else if ext == "" {
			manager.Warningf("Artwork %d cannot be identified", i+1)
			continue
		}

		// Modify the filename if there is more than one artwork
		filename := filepath.Base(cmd.Path) + ext
		if len(artworks) > 1 {
			filename = filepath.Base(cmd.Path) + fmt.Sprintf("%d.%s", i+1, ext)
		}

		// Write the file
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		if _, err := w.Write(data); err != nil {
			return err
		}
		fmt.Println("Written ", filename)
	}

	return nil
}
