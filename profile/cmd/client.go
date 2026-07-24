package cmd

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	httpclient "github.com/mutablelogic/go-media/profile/httpclient"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientCommands struct {
	ClientCodecCommands
	ClientPixelFormatCommands
	ClientProfileCommands
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func withClient(cmd server.Cmd, name string, fn func(context.Context, *httpclient.Client) error) (err error) {
	endpoint, opts, err := cmd.ClientEndpoint()
	if err != nil {
		return err
	}
	client, err := httpclient.New(endpoint, opts...)
	if err != nil {
		return err
	}
	ctx, endSpan := otel.StartSpan(cmd.Tracer(), cmd.Context(), name)
	defer func() { endSpan(err) }()
	return fn(ctx, client)
}
