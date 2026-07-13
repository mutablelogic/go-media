package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Packages

	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	metadata "github.com/mutablelogic/go-media/metadata"
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
	EncodingCLICommands
}

type BaseCmd struct {
	ChromaprintKey string `name:"chromaprint-key" env:"CHROMAPRINT_KEY" help:"AcoustID API key for chromaprint lookups"`
}

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
	Path      string   `arg:"" name:"path" type:"path" help:"File to extract metadata from." default:"."`
	Out       string   `flag:"" name:"out" help:"Output template for metadata files."`
	Recursive bool     `flag:"" name:"recursive" short:"r" help:"Recursively extract metadata from files in a directory." negatable:""`
	Exclude   []string `flag:"" name:"exclude" help:"Exclude files with these extensions (e.g. .jpg, .png)."`
	Namespace string   `flag:"" name:"namespace" help:"Namespace to extract." default:""`
}

type ArtworkCmd struct {
	BaseCmd
	File string `arg:"" name:"path" type:"path" help:"File to extract metadata from."`
	Out  string `flag:"" name:"out" help:"Output template for artwork files." default:"{name}_{key}{ext}"`
}

type ProbeCmd struct {
	BaseCmd
	File string `arg:"" name:"path" type:"path" help:"File to probe."`
	schema.ProbeRequest
}

func (c *MetadataCmd) Run(ctx server.Cmd) error {
	//	json, termwidth := c.IsJSONOutput(ctx)
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

		// Walk the filepath
		log := ctx.Logger()
		return WalkFS(ctx.Context(), os.DirFS(c.Path), func(ctx context.Context, path string, info fs.DirEntry, tmpl *Templater) error {
			// Allow recusive walking of directories, but only process files
			if info.IsDir() {
				return nil
			}

			// Only open regular files
			if !info.Type().IsRegular() {
				log.WarnContext(ctx, "Skipping non-regular file", "path", path)
				return nil
			}

			// Open the file
			r, err := os.Open(filepath.Join(c.Path, path))
			if err != nil {
				return err
			}
			defer r.Close()

			// Read the metadata from the file, logging any warnings but not failing on them
			var warn error
			meta, err := manager.GetMetadata(ctx, r, ns, &warn)
			if errors.Is(err, gomedia.ErrNotImplemented) {
				log.WarnContext(ctx, "Skipping unsupported file", "path", path, "error", err.Error())
				return nil
			} else if err != nil {
				return err
			}
			if warn != nil {
				log.WarnContext(ctx, "Warning reading metadata", "path", path, "error", warn.Error())
			}

			if tmpl != nil {
				if out, err := tmpl.Path(map[string]any{
					"path": path,
					"name": info.Name(),
				}); err == nil {
					fmt.Println(out, "=>", types.Stringify(meta))
				} else {
					return err
				}
			} else {
				fmt.Println(types.Stringify(meta))
			}
			return nil
		}, opts...)
		/*


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
		*/
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

///////////////////////////////////////////////////////////////////////////////
// ENCODING

type EncodingCLICommands struct {
	AudioSegment AudioSegmentCmd `cmd:"" name:"audio-segment" help:"Segment audio and log segments." group:"ENCODING"`
}

type AudioSegmentCmd struct {
	BaseCmd
	File             string        `arg:"" name:"file" type:"existingfile" help:"File to segment."`
	Out              string        `flag:"" name:"out" help:"Output directory for encoded segment M4A files." default:"." type:"path"`
	Duration         time.Duration `flag:"" name:"duration" help:"Target segment duration (e.g. 30s). Use 0s to disable fixed-size splits."`
	Silence          bool          `flag:"" name:"silence" help:"Enable silence-based segmentation." negatable:"" default:"true"`
	SilenceDuration  time.Duration `flag:"" name:"silence-duration" help:"Minimum silence duration for silence-based splitting (e.g. 500ms). Also enables silence splitting." default:"0s"`
	SilenceThreshold float64       `flag:"" name:"silence-threshold" help:"Silence threshold as RMS energy (0.0-1.0). Also enables silence splitting. 0 uses auto threshold (0.005)." default:"0"`
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

func (c *AudioSegmentCmd) Run(ctx server.Cmd) error {
	return c.WithManager(ctx, func(manager *manager.Media) error {
		r, err := os.Open(c.File)
		if err != nil {
			return err
		}
		defer r.Close()

		return manager.SegmentAudio(ctx.Context(), schema.SegmentAudioRequest{
			Reader:           r,
			OutputDir:        c.Out,
			Duration:         c.Duration,
			Silence:          c.Silence,
			SilenceDuration:  c.SilenceDuration,
			SilenceThreshold: c.SilenceThreshold,
		})
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (runner *BaseCmd) IsJSONOutput(ctx server.Cmd) (bool, int) {
	width := ctx.IsTerm()
	return ctx.IsDebug() || width == 0, width
}

func (runner *BaseCmd) WithManager(ctx server.Cmd, fn func(*manager.Media) error) error {
	// Client opts
	_, clientopts, err := ctx.ClientEndpoint()
	if err != nil {
		return err
	}

	// Set basic mamager options
	opts := []manager.Opt{
		manager.WithTracer(ctx.Tracer()),
	}

	// Chromaprint key
	if runner.ChromaprintKey != "" {
		opts = append(opts, manager.WithAcoustIDKey(runner.ChromaprintKey, clientopts...))
	}

	// Create a manager and then call the function with the manager, returning any error
	if manager, err := manager.New(ctx.Context(), opts...); err != nil {
		return err
	} else {
		return fn(manager)
	}
}
