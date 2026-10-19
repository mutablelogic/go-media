package tmdb

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-media/tmdb/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// GetMovie returns the details of a movie, such as its runtime
func (c *Client) GetMovie(ctx context.Context, id uint64) (*schema.MovieDetails, error) {
	var result schema.MovieDetails
	if err := c.DoWithContext(ctx, nil, &result, client.OptPath("movie", id)); err != nil {
		return nil, err
	}
	return types.Ptr(result), nil
}
