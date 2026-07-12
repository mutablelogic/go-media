package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	metadata "github.com/mutablelogic/go-media/metadata"
	xmp "github.com/mutablelogic/go-media/pkg/xmp"
	server "github.com/mutablelogic/go-server"
	tui "github.com/mutablelogic/go-server/pkg/tui"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Imports
	_ "github.com/mutablelogic/go-media/metadata/application"
	_ "github.com/mutablelogic/go-media/metadata/audio"
	_ "github.com/mutablelogic/go-media/metadata/image"
	_ "github.com/mutablelogic/go-media/metadata/video"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLICommands struct {
	MetadataCLICommands
	CapabilitiesCLICommands
}

type BaseCmd struct{}

///////////////////////////////////////////////////////////////////////////////
// METADATA

type MetadataCLICommands struct {
	Metadata MetadataCmd `cmd:"" name:"metadata" help:"Extract metadata." group:"METADATA"`
	Artwork  ArtworkCmd  `cmd:"" name:"artwork" help:"Extract artwork." group:"METADATA"`
	Probe    ProbeCmd    `cmd:"" name:"probe" help:"Probe media file." group:"METADATA"`
	MetadataChromaprintCLICommands
}

type MetadataCmd struct {
	BaseCmd
	File      string `arg:"" name:"file" type:"existingfile" help:"File to extract metadata from."`
	Namespace string `flag:"" name:"namespace" help:"Namespace to extract." default:""`
	XMP       bool   `flag:"" name:"xmp" help:"Output metadata in XMP format." negatable:""`
}

type ArtworkCmd struct {
	BaseCmd
	File string `arg:"" name:"file" type:"existingfile" help:"File to extract metadata from."`
	Out  string `flag:"" name:"out" help:"Output template for artwork files." default:"{name}_{key}{ext}"`
}

type ProbeCmd struct {
	BaseCmd
	File string `arg:"" name:"file" type:"existingfile" help:"File to probe."`
	schema.ProbeRequest
}

func (c *MetadataCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		// Convert namespace
		var ns string
		if ns = strings.TrimSpace(c.Namespace); ns != "" {
			if !types.IsIdentifier(ns) {
				return fmt.Errorf("invalid namespace: %q", ns)
			} else {
				ns = ns + ":"
			}
		}

		// Open the file
		r, err := os.Open(c.File)
		if err != nil {
			return err
		}
		defer r.Close()

		var warn error
		meta, err := manager.GetMetadata(ctx.Context(), r, ns, &warn)
		if err != nil {
			return err
		}
		if warn != nil {
			fmt.Fprintln(os.Stderr, "Warning:", warn)
		}

		// If the XMP flag is set, output the metadata in XMP format
		if c.XMP {
			metadataItems := make([]gomedia.Metadata, 0, len(meta.Meta))
			for _, item := range meta.Meta {
				if item.Metadata != nil {
					metadataItems = append(metadataItems, item.Metadata)
				}
			}
			fmt.Println(xmp.FromMetadata(metadataItems).String())
			return nil
		}

		if json {
			fmt.Println(types.Stringify(meta))
			return nil
		}

		// Output a table to the terminal
		table := tui.TableFor[schema.MetaItem](tui.SetWidth(termwidth))
		if _, err := table.Write(os.Stdout, meta.Meta...); err != nil {
			return err
		}
		return nil
	})
}

func (c *ArtworkCmd) Run(ctx server.Cmd) error {
	return c.WithManager(ctx, func(manager *manager.Media) error {
		// Open the file
		r, err := os.Open(c.File)
		if err != nil {
			return err
		}
		defer r.Close()

		// Create the basic arguments for the output template
		dir, base := filepath.Split(c.File)
		args := []any{
			"dir", dir,
			"base", base,
			"name", strings.TrimSuffix(base, filepath.Ext(c.File)),
		}

		// Extract the artwork from the file
		var warn error
		meta, err := manager.GetMetadata(ctx.Context(), r, "artwork:", &warn)
		if err != nil {
			return err
		}
		if warn != nil {
			fmt.Fprintln(os.Stderr, "Warning:", warn)
		}

		for i, item := range meta.Meta {
			_, name := func() (string, string) {
				parts := strings.SplitN(item.Key(), ":", 2)
				if len(parts) != 2 {
					return "", ""
				}
				return parts[0], parts[1]
			}()
			if image := item.Bytes(); len(image) == 0 {
				continue
			} else if mimetype := item.Value(); mimetype == "" {
				continue
			} else if ext := metadata.ExtensionByType(mimetype); ext == "" {
				continue
			} else if ext != "" {
				out := PathFromTemplate(c.Out, append(args, "index", i, "key", item.Key(), "ext", ext, "key", name)...)
				base := filepath.Dir(out)
				if err := os.MkdirAll(base, 0755); err != nil {
					return err
				} else if f, err := os.Create(out); err != nil {
					return err
				} else if _, err := f.Write(image); err != nil {
					f.Close()
					return err
				} else if err := f.Close(); err != nil {
					return err
				}
				fmt.Println("Extracting artwork:", item.Key(), "=>", out)
			}
		}

		return nil
	})
}

func (c *ProbeCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		// Open the file
		r, err := os.Open(c.File)
		if err != nil {
			return err
		}
		defer r.Close()

		// Probe the media file
		resp, err := manager.Probe(ctx.Context(), schema.ProbeRequest{
			Reader:      r,
			InputFormat: c.InputFormat,
			InputOpts:   c.InputOpts,
		})
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		fmt.Printf("Format: %s\n", resp.Format)
		if resp.Description != "" {
			fmt.Printf("Description: %s\n", resp.Description)
		}
		fmt.Printf("Duration: %.3fs\n", resp.Duration)
		if len(resp.MimeTypes) > 0 {
			fmt.Printf("MIME Types: %s\n", strings.Join(resp.MimeTypes, ", "))
		}

		rows := make([]schema.Stream, 0, len(resp.Streams))
		for _, s := range resp.Streams {
			if s == nil {
				continue
			}
			rows = append(rows, *s)
		}

		if len(rows) > 0 {
			table := tui.TableFor[schema.Stream](tui.SetWidth(termwidth))
			if _, err := table.Write(os.Stdout, rows...); err != nil {
				return err
			}
		}
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// CAPABILITIES

type CapabilitiesCLICommands struct {
	AudioChannels AudioChannelsCmd `cmd:"" name:"audio-channels" help:"List audio channel layouts." group:"CAPABILITIES"`
	Codecs        CodecCmd         `cmd:"" name:"codecs" help:"List codecs." group:"CAPABILITIES"`
	Filters       FiltersCmd       `cmd:"" name:"filters" help:"List filters." group:"CAPABILITIES"`
	Formats       FormatsCmd       `cmd:"" name:"formats" help:"List formats and devices." group:"CAPABILITIES"`
	PixelFormats  PixelFormatsCmd  `cmd:"" name:"pixel-formats" help:"List pixel formats." group:"CAPABILITIES"`
	SampleFormats SampleFormatsCmd `cmd:"" name:"sample-formats" help:"List sample formats." group:"CAPABILITIES"`
}

type AudioChannelsCmd struct {
	BaseCmd
	schema.ListAudioChannelLayoutRequest
}

type CodecCmd struct {
	BaseCmd
	schema.ListCodecRequest
}

type FormatsCmd struct {
	BaseCmd
	schema.ListFormatRequest
}

type FiltersCmd struct {
	BaseCmd
	schema.ListFilterRequest
}

type PixelFormatsCmd struct {
	BaseCmd
	schema.ListPixelFormatRequest
}

type SampleFormatsCmd struct {
	BaseCmd
	schema.ListSampleFormatRequest
}

func (c *AudioChannelsCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		resp, err := manager.ListAudioChannelLayouts(ctx.Context(), c.ListAudioChannelLayoutRequest)
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		table := tui.TableFor[schema.AudioChannelLayout](tui.SetWidth(termwidth))
		if _, err := table.Write(os.Stdout, resp...); err != nil {
			return err
		}
		return nil
	})
}

func (c *CodecCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		resp, err := manager.ListCodecs(ctx.Context(), c.ListCodecRequest)
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		table := tui.TableFor[schema.Codec](tui.SetWidth(termwidth))
		if _, err := table.Write(os.Stdout, resp...); err != nil {
			return err
		}
		return nil
	})
}

func (c *FiltersCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		resp, err := manager.ListFilters(ctx.Context(), c.ListFilterRequest)
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		table := tui.TableFor[schema.Filter](tui.SetWidth(termwidth))
		if _, err := table.Write(os.Stdout, resp...); err != nil {
			return err
		}
		return nil
	})
}

func (c *FormatsCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		resp, err := manager.ListFormats(ctx.Context(), c.ListFormatRequest)
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		table := tui.TableFor[schema.Format](tui.SetWidth(termwidth))
		if _, err := table.Write(os.Stdout, resp...); err != nil {
			return err
		}

		deviceTable := tui.TableFor[schema.Device](tui.SetWidth(termwidth))
		for _, format := range resp {
			if len(format.Devices) == 0 {
				continue
			}

			fmt.Printf("\nDevices for %s:\n", format.Name)
			if _, err := deviceTable.Write(os.Stdout, format.Devices...); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *PixelFormatsCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		resp, err := manager.ListPixelFormats(ctx.Context(), c.ListPixelFormatRequest)
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		table := tui.TableFor[schema.PixelFormat](tui.SetWidth(termwidth))
		if _, err := table.Write(os.Stdout, resp...); err != nil {
			return err
		}
		return nil
	})
}

func (c *SampleFormatsCmd) Run(ctx server.Cmd) error {
	json, termwidth := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		resp, err := manager.ListSampleFormats(ctx.Context(), c.ListSampleFormatRequest)
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		table := tui.TableFor[schema.SampleFormat](tui.SetWidth(termwidth))
		if _, err := table.Write(os.Stdout, resp...); err != nil {
			return err
		}
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (runner *BaseCmd) IsJSONOutput(ctx server.Cmd) (bool, int) {
	width := ctx.IsTerm()
	return ctx.IsDebug() || width == 0, width
}

func (runner *BaseCmd) WithManager(ctx server.Cmd, fn func(*manager.Media) error) error {
	// Set basic mamager options
	opts := []manager.Opt{
		manager.WithTracer(ctx.Tracer()),
	}

	// Create a manager and then call the function with the manager, returning any error
	if manager, err := manager.New(ctx.Context(), opts...); err != nil {
		return err
	} else {
		return fn(manager)
	}
}
