package tmdb

import (
	"context"
	"net/url"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-media/tmdb/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// withExternalIDs requests a show's or episode's identifiers in other
// databases, such as IMDb, with its details, rather than separately
var withExternalIDs = url.Values{"append_to_response": []string{"external_ids"}}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// GetTVShow returns the details of a TV show, such as the typical runtimes
// of its episodes, with its identifiers in other databases, such as IMDb
func (c *Client) GetTVShow(ctx context.Context, id uint64) (*schema.TVShowDetails, error) {
	var result schema.TVShowDetails
	if err := c.DoWithContext(ctx, nil, &result, client.OptPath("tv", id), client.OptQuery(withExternalIDs)); err != nil {
		return nil, err
	}
	return types.Ptr(result), nil
}

// GetTVEpisode returns an episode of a TV show, by its season and episode
// numbers, such as its name and runtime, with its identifiers in other
// databases, such as IMDb
func (c *Client) GetTVEpisode(ctx context.Context, id, season, episode uint64) (*schema.TVEpisode, error) {
	var result schema.TVEpisode
	if err := c.DoWithContext(ctx, nil, &result, client.OptPath("tv", id, "season", season, "episode", episode), client.OptQuery(withExternalIDs)); err != nil {
		return nil, err
	}
	return types.Ptr(result), nil
}
