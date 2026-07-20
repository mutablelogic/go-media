package cmd

import (
	"context"
	"fmt"

	// Packages
	httpclient "github.com/mutablelogic/go-media/profile/httpclient"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ListCodecs struct {
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListCodecs) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListCodecs", func(ctx context.Context, client *httpclient.Client) error {
		// List the codecs
		codecs, err := client.ListCodecs(ctx)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(codecs))

		// Return success
		return nil
	})
}
