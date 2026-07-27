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

type ClientChannelLayoutCommands struct {
	ListChannelLayouts ListChannelLayouts `cmd:"" name:"channellayouts" help:"List the available channel layouts." group:"CAPABILITIES"`
}

type ListChannelLayouts struct {
	schema.ChannelLayoutListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListChannelLayouts) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListChannelLayouts", func(ctx context.Context, client *httpclient.Client) error {
		// List the channel layouts
		channellayouts, err := client.ListChannelLayouts(ctx, cmd.ChannelLayoutListRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(channellayouts))

		// Return success
		return nil
	})
}
