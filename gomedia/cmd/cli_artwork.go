package cmd

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	metadata "github.com/mutablelogic/go-media/metadata"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ArtworkCmd struct {
	BaseCmd
	Path      string   `arg:"" name:"path" type:"path" help:"File to extract artwork from." default:"."`
	Out       string   `flag:"" name:"out" help:"Output template for artwork files." required:""`
	Recursive bool     `flag:"" name:"recursive" short:"r" help:"Recursively extract artwork from files in a directory." negatable:""`
	Exclude   []string `flag:"" name:"exclude" help:"Exclude files with these extensions (e.g. .jpg, .png)."`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *ArtworkCmd) Run(ctx server.Cmd) error {
	log := ctx.Logger()

	// Gather FS walking options
	opts := []WalkOpt{}
	if c.Recursive {
		opts = append(opts, WithRecursive())
	}
	if len(c.Exclude) > 0 {
		opts = append(opts, WithExcludeExt(c.Exclude...))
	}
	if c.Out != "" {
		opts = append(opts, WithTemplate(c.Out))
	}

	return c.WithManager(ctx, func(manager *manager.Media) error {
		return WalkFS(ctx.Context(), c.Path, func(ctx context.Context, fullPath string, relPath string, entry fs.DirEntry, tmpl *Templater) error {
			// Skip directories, but allow the walk to continue into them
			if entry.IsDir() {
				return nil
			}

			// Only open regular files
			if !entry.Type().IsRegular() {
				log.WarnContext(ctx, "Skipping non-regular file", "path", relPath)
				return nil
			}

			// Open the file
			r, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer r.Close()

			// Read the metadata from the file, logging any warnings but not failing on them
			var warn error
			meta, err := manager.GetMetadata(ctx, r, "artwork:", &warn)
			if errors.Is(err, gomedia.ErrNotImplemented) {
				log.WarnContext(ctx, "Skipping unsupported file", "path", relPath, "error", err.Error())
				return nil
			} else if err != nil {
				return err
			}
			if warn != nil {
				log.WarnContext(ctx, "Warning reading metadata", "path", relPath, "error", warn.Error())
			}

			if tmpl == nil {
				return gomedia.ErrBadParameter.With("Missing -out parameter")
			}

			for index, item := range meta.Meta {
				// If there's no data, skip it
				if len(item.Bytes()) == 0 {
					continue
				}

				// Extract the mimetype of the artwork
				tmpl.Create(map[string]any{
					"index": index,
					"path":  relPath,
					"name":  entry.Name(),
					"type":  item.Value(),
					"ext":   metadata.ExtensionByType(item.Value()),
					"key":   strings.TrimPrefix(item.Key(), "artwork:"),
				}, func(w io.Writer) error {
					if w, ok := w.(gomedia.NamedWriter); ok {
						log.InfoContext(ctx, "Writing artwork file "+filepath.Base(w.Name()), "path", w.Name())
					}

					// Write the artwork data to the output file
					_, err := w.Write(item.Bytes())
					return err
				})

			}

			return nil
		}, opts...)
	})
}
