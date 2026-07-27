package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	xmp "github.com/mutablelogic/go-media/pkg/xmp"
	server "github.com/mutablelogic/go-server"
	tui "github.com/mutablelogic/go-server/pkg/tui"
	types "github.com/mutablelogic/go-server/pkg/types"
	yaml "gopkg.in/yaml.v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataCmd struct {
	BaseCmd
	Path      string   `arg:"" name:"path" type:"path" help:"File to extract metadata from." default:"."`
	Out       string   `flag:"" name:"out" help:"Output template for metadata files."`
	Recursive bool     `flag:"" name:"recursive" short:"r" help:"Recursively extract metadata from files in a directory." negatable:""`
	Exclude   []string `flag:"" name:"exclude" help:"Exclude files with these extensions (e.g. .jpg, .png)."`
	Namespace string   `flag:"" name:"namespace" help:"Namespace to extract." default:""`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *MetadataCmd) Run(ctx server.Cmd) error {
	//	json, termwidth := c.IsJSONOutput(ctx)
	log := ctx.Logger()
	stdout := true
	table := tui.TableFor[schema.MetaItem](tui.SetWidth(ctx.IsTerm()))

	// Convert namespace
	var ns string
	if ns = strings.TrimSpace(c.Namespace); ns != "" {
		if !types.IsIdentifier(ns) {
			return fmt.Errorf("invalid namespace: %q", ns)
		} else {
			ns = ns + ":"
		}
	}

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
		stdout = false
	}

	return c.WithManager(ctx, func(manager *manager.Media) error {
		if err := WalkFS(ctx.Context(), c.Path, func(ctx context.Context, fullPath string, relPath string, entry fs.DirEntry, tmpl *Templater) error {
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
			meta, err := manager.GetMetadata(ctx, r, ns, &warn)
			if errors.Is(err, gomedia.ErrNotImplemented) {
				log.WarnContext(ctx, "Skipping unsupported file", "path", relPath, "error", err.Error())
				return nil
			} else if err != nil {
				return err
			}
			if warn != nil {
				log.WarnContext(ctx, "Warning reading metadata", "path", relPath, "error", warn.Error())
			}

			// Write according to a template
			if tmpl != nil {
				return tmpl.Create(map[string]any{"path": relPath, "name": entry.Name()}, func(w io.Writer) error {
					if w, ok := w.(gomedia.NamedWriter); ok {
						log.InfoContext(ctx, "Writing metadata file "+filepath.Base(w.Name()), "path", w.Name())
						switch ext := strings.ToLower(filepath.Ext(w.Name())); ext {
						case ".xmp":
							metadataItems := make([]gomedia.Metadata, 0, len(meta.Meta))
							for _, item := range meta.Meta {
								if item.Metadata != nil {
									metadataItems = append(metadataItems, item.Metadata)
								}
							}
							return xmp.FromMetadata(metadataItems).Write(w)
						case ".json":
							enc := json.NewEncoder(w)
							enc.SetIndent("", "  ")
							return enc.Encode(meta)
						case ".yml", ".yaml":
							enc := yaml.NewEncoder(w)
							enc.SetIndent(2)
							return enc.Encode(meta)
						default:
							return gomedia.ErrBadParameter.With("unsupported output file extension: " + ext)
						}
					}
					return nil
				})
			}

			// Write according to a table
			if len(meta.Meta) > 0 {
				meta.Meta[0].Name = relPath
				table.Append(meta.Meta...)
			}

			return nil
		}, opts...); err != nil {
			return err
		}

		// Write out the table
		if stdout && table != nil {
			if _, err := table.Write(os.Stdout); err != nil {
				return err
			}
		}

		// Return success
		return nil
	})
}
