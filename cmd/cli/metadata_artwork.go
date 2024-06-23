package main

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-media"
)

type MetadataCmd struct {
	Path string `arg:"" required:"" help:"Media file" type:"path"`
}

type ArtworkCmd struct {
	Path string `arg:"" required:"" help:"Media file" type:"path"`
}

func (cmd *MetadataCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	reader, err := manager.Open(cmd.Path, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Print metadata
	opts := []tablewriter.TableOpt{
		tablewriter.OptHeader(),
		tablewriter.OptOutputText(),
	}
	return tablewriter.New(os.Stdout, opts...).Write(reader.Metadata())
}

func (cmd *ArtworkCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	reader, err := manager.Open(cmd.Path, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	artwork := reader.Metadata(media.MetaArtwork)
	if len(artwork) == 0 {
		return errors.New("no artwork")
	}

	for i, a := range artwork {
		data := a.Value().([]byte)
		mimetype := http.DetectContentType(data)
		if !strings.HasPrefix(mimetype, "image/") {
			return fmt.Errorf("invalid mimetype %q for stream %d", mimetype, i)
		}
		ext, err := mime.ExtensionsByType(mimetype)
		if err != nil {
			return err
		}
		if len(ext) == 0 {
			return fmt.Errorf("no extension for mimetype %q", mimetype)
		}

		// Use last extension
		filename := filepath.Base(cmd.Path) + ext[len(ext)-1]
		if len(artwork) > 1 {
			filename = filepath.Base(cmd.Path) + fmt.Sprintf("%d.%s", i+1, ext[len(ext)-1])
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
