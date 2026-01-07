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
	r, err := client.NewJSONRequest(req)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.ListAudioChannelLayoutResponse
	if err := c.DoWithContext(ctx, r, &response, client.OptPath("audiochannellayout")); err != nil {
		return nil, err
	}

	// Return the response
	return response, nil
}
