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

type ClientSampleFormatCommands struct {
	ListSampleFormats ListSampleFormats `cmd:"" name:"sampleformats" help:"List the available sample formats." group:"CLIENT"`
}

type ListSampleFormats struct {
	schema.SampleFormatListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListSampleFormats) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListSampleFormats", func(ctx context.Context, client *httpclient.Client) error {
		// List the sample formats
		sampleformats, err := client.ListSampleFormats(ctx, cmd.SampleFormatListRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(sampleformats))

		// Return success
		return nil
	})
}
