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

type ClientFormatCommands struct {
	ListContainerFormats ListContainerFormats `cmd:"" name:"formats" help:"List the available container formats." group:"CLIENT"`
}

type ListContainerFormats struct {
	schema.FormatListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListContainerFormats) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListContainerFormats", func(ctx context.Context, client *httpclient.Client) error {
		// List the container formats
		formats, err := client.ListContainerFormats(ctx, cmd.FormatListRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(formats))

		// Return success
		return nil
	})
}
