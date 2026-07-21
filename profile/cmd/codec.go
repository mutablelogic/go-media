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

type ClientCodecCommands struct {
	ListCodecs ListCodecs `cmd:"" name:"codecs" help:"List the available codecs." group:"CLIENT"`
	GetCodec   GetCodec   `cmd:"" name:"codec" help:"Get the details of a codec." group:"CLIENT"`
}

type GetCodec struct {
	Name string `arg:"" name:"name" help:"Name of the codec."`
}

type ListCodecs struct {
	schema.CodecListRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *GetCodec) Run(ctx server.Cmd) error {
	return withClient(ctx, "GetCodec", func(ctx context.Context, client *httpclient.Client) error {
		// Get the codec
		codec, err := client.GetCodec(ctx, cmd.Name)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(codec))

		// Return success
		return nil
	})
}

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
