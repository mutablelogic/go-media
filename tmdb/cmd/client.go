package cmd

import (
	"context"

	// Packages
	kong "github.com/alecthomas/kong"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	httpclient "github.com/mutablelogic/go-media/tmdb/httpclient"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientCommands struct {
	Token string `flag:"" name:"tmdb-token" env:"TMDB_TOKEN" help:"TMDB API Read Access Token."`
	SearchCommands
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// AfterApply binds the ClientCommands instance into the kong context so that
// nested command Run() methods can request it (and its APIKey) as an argument.
func (c *ClientCommands) AfterApply(ctx *kong.Context) error {
	ctx.Bind(c)
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func withClient(cmd server.Cmd, token string, name string, fn func(context.Context, *httpclient.Client) error) (err error) {
	_, opts, err := cmd.ClientEndpoint()
	if err != nil {
		return err
	}

	client, err := httpclient.New(token, opts...)
	if err != nil {
		return err
	}
	ctx, endSpan := otel.StartSpan(cmd.Tracer(), cmd.Context(), name)
	defer func() { endSpan(err) }()
	return fn(ctx, client)
}
