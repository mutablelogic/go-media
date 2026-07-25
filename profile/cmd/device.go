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

type ClientDeviceCommands struct {
	ListDevices ListDevices `cmd:"" name:"devices" help:"List the available devices." group:"CAPABILITIES"`
}

type ListDevices struct {
	schema.DeviceListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListDevices) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListDevices", func(ctx context.Context, client *httpclient.Client) error {
		// List the devices
		devices, err := client.ListDevices(ctx, cmd.DeviceListRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(devices))

		// Return success
		return nil
	})
}
