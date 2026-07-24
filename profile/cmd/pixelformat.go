package cmd

import (
	"context"
	"fmt"

	// Packages
	httpclient "github.com/mutablelogic/go-media/profile/httpclient"
	schema "github.com/mutablelogic/go-media/profile/schema"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientPixelFormatCommands struct {
	ListPixelFormats ListPixelFormats `cmd:"" name:"pixelformats" help:"List the available pixel formats." group:"CLIENT"`
}

type ListPixelFormats struct {
	schema.PixelFormatListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListPixelFormats) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListPixelFormats", func(ctx context.Context, client *httpclient.Client) error {
		// List the pixel formats
		pixelformats, err := client.ListPixelFormats(ctx, cmd.PixelFormatListRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(pixelformats))

		// Return success
		return nil
	})
}
