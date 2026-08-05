package cmd

import (
	"context"
	"fmt"

	// Packages
	httpclient "github.com/mutablelogic/go-media/tmdb/httpclient"
	schema "github.com/mutablelogic/go-media/tmdb/schema"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SearchCommands struct {
	Movies SearchMoviesCmd `cmd:"" name:"search-movies" help:"Search for movies." group:"SEARCH"`
}

type SearchMoviesCmd struct {
	schema.MovieSearchRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *SearchMoviesCmd) Run(ctx server.Cmd, client *ClientCommands) error {
	logger := ctx.Logger()
	return withClient(ctx, client.Token, "SearchMovies", func(ctx context.Context, c *httpclient.Client) error {
		logger.InfoContext(ctx, "Performing Search", "req", types.Stringify(cmd.MovieSearchRequest))
		resp, err := c.SearchMovies(ctx, &cmd.MovieSearchRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(resp))
		return nil
	})
}
