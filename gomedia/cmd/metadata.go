package cmd

import (
	"context"
	"fmt"
	"os"

	// Packages
	httpclient "github.com/mutablelogic/go-media/gomedia/httpclient"
	task "github.com/mutablelogic/go-media/gomedia/task"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataCommands struct {
	Probe ProbeCmd `cmd:"" name:"probe" help:"Probe a media file and return its container format, streams, and metadata." group:"METADATA"`
}

type ProbeCmd struct {
	Path   string   `arg:"" name:"path" type:"existingfile" help:"Path to the media file to probe."`
	Format string   `name:"format" help:"Input format name (e.g. mpegts)."`
	Opts   []string `name:"opts" help:"Input format options."`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ProbeCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "Probe", func(ctx context.Context, client *httpclient.Client) error {
		f, err := os.Open(cmd.Path)
		if err != nil {
			return err
		}
		defer f.Close()

		req := task.ProbeRequest{
			Reader: f,
			Format: cmd.Format,
			Opts:   cmd.Opts,
		}

		// Upload as multipart/form-data rather than a raw body: only the
		// form-data path carries the file's name through to the server (a
		// raw request body has no place to put a filename), and *os.File
		// implements gomedia.NamedReader, so this is what actually shows
		// up as ProbeResponse.Name.
		response, err := client.Probe(ctx, req, types.ContentTypeFormData, nil)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(response))
		return nil
	})
}
