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

func (c *Client) SearchMovies(ctx context.Context, req schema.MovieSearchRequest) (*schema.MovieSearchResponse, error) {
	var result schema.MovieSearchResponse
	if err := c.DoWithContext(ctx, nil, &result, client.OptPath("search", "movie"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}
	return types.Ptr(result), nil
}

// SearchTVShows returns a page of the TV shows whose names match the search,
// with the most relevant first. TMDB searches shows, rather than episodes,
// which are got from a show with GetTVEpisode.
func (c *Client) SearchTVShows(ctx context.Context, req schema.TVSearchRequest) (*schema.TVSearchResponse, error) {
	var result schema.TVSearchResponse
	if err := c.DoWithContext(ctx, nil, &result, client.OptPath("search", "tv"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}
	return types.Ptr(result), nil
}
