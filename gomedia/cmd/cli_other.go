package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	// Packages
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	server "github.com/mutablelogic/go-server"
	tui "github.com/mutablelogic/go-server/pkg/tui"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ProbeCmd struct {
	BaseCmd
	File string `arg:"" name:"path" type:"path" help:"File to probe."`
	schema.ProbeRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

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
