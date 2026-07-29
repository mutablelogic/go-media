package cmd

import (
	"context"
	"fmt"
	"os"

	// Packages
	httpclient "github.com/mutablelogic/go-media/gomedia/httpclient"
	task "github.com/mutablelogic/go-media/gomedia/task"
	taskmetadata "github.com/mutablelogic/go-media/task/metadata"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataCommands struct {
	ProbeMedia  ProbeMediaCmd  `cmd:"" name:"probe-media" help:"Probe a media file and return its container format, streams, and metadata." group:"METADATA"`
	ProbeSource ProbeSourceCmd `cmd:"" name:"probe-source" help:"Probe a URL or device and return its container format, streams, and metadata." group:"METADATA"`
	Metadata    MetadataCmd    `cmd:"" name:"metadata" help:"Extract container-level metadata from a media file." group:"METADATA"`
}

type ProbeMediaCmd struct {
	Path   string   `arg:"" name:"path" type:"existingfile" help:"Path to the media file to probe."`
	Format string   `name:"format" help:"Input format name (e.g. mpegts)."`
	Opts   []string `name:"opts" help:"Input format options."`
}

type ProbeSourceCmd struct {
	Url    string   `arg:"" name:"url" help:"URL to probe, e.g. \"https://example.com/sample.mp3\" or \"device://avfoundation/0:0\"."`
	Format string   `name:"format" help:"Input format name (e.g. mpegts)."`
	Opts   []string `name:"opts" help:"Input format options."`
}

type MetadataCmd struct {
	Path string `arg:"" name:"path" type:"existingfile" help:"Path to the media file to extract metadata from."`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ProbeMediaCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "ProbeMedia", func(ctx context.Context, client *httpclient.Client) error {
		f, err := os.Open(cmd.Path)
		if err != nil {
			return err
		}
		defer f.Close()

		req := task.ProbeMediaRequest{
			Reader: f,
			ProbeRequestOpts: task.ProbeRequestOpts{
				Format: cmd.Format,
				Opts:   cmd.Opts,
			},
		}

		// Upload as multipart/form-data rather than a raw body: only the
		// form-data path carries the file's name through to the server (a
		// raw request body has no place to put a filename), and *os.File
		// implements gomedia.NamedReader, so this is what actually shows
		// up as ProbeResponse.Name.
		response, err := client.ProbeMedia(ctx, req, types.ContentTypeFormData, nil)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(response))
		return nil
	})
}

func (cmd *ProbeSourceCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "ProbeSource", func(ctx context.Context, client *httpclient.Client) error {
		req := task.ProbeSourceRequest{
			Url: cmd.Url,
			ProbeRequestOpts: task.ProbeRequestOpts{
				Format: cmd.Format,
				Opts:   cmd.Opts,
			},
		}

		response, err := client.ProbeSource(ctx, req)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(response))
		return nil
	})
}

func (cmd *MetadataCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "Metadata", func(ctx context.Context, client *httpclient.Client) error {
		f, err := os.Open(cmd.Path)
		if err != nil {
			return err
		}
		defer f.Close()

		// Upload as multipart/form-data
		response, err := client.Metadata(ctx, taskmetadata.MetadataRequest{
			Reader: f,
		}, types.ContentTypeFormData, nil)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(response))
		return nil
	})
}
