package httpclient

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListAudioChannelLayouts(ctx context.Context, req *schema.ListAudioChannelLayoutRequest) (schema.ListAudioChannelLayoutResponse, error) {
	// Construct the query
	opts := []client.RequestOpt{
		client.OptPath("audiochannellayout"),
		client.OptQuery(req.QueryValues()),
	}

	// Perform request
	var response schema.ListAudioChannelLayoutResponse
	if err := c.DoWithContext(ctx, client.MethodGet, &response, opts...); err != nil {
		return nil, err
	}

	// Return the response
	return response, nil
}

func (c *Client) ListCodecs(ctx context.Context, req *schema.ListCodecRequest) (schema.ListCodecResponse, error) {
	// Construct the query
	opts := []client.RequestOpt{
		client.OptPath("codec"),
		client.OptQuery(req.QueryValues()),
	}

	// Perform request
	var response schema.ListCodecResponse
	if err := c.DoWithContext(ctx, client.MethodGet, &response, opts...); err != nil {
		return nil, err
	}

	// Return the response
	return response, nil
}
