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

type ListCodecs struct {
	schema.CodecListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListCodecs) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListCodecs", func(ctx context.Context, client *httpclient.Client) error {
		// List the codecs
		codecs, err := client.ListCodecs(ctx, cmd.CodecListRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(codecs))

		// Return success
		return nil
	})
}
