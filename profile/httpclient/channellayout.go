package httpclient

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-media/profile/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *Client) ListChannelLayouts(ctx context.Context, req schema.ChannelLayoutListRequest) (*schema.ChannelLayoutList, error) {
	// Perform request
	var response schema.ChannelLayoutList
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("channellayout"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
