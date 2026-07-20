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

func (c *Client) ListCodecs(ctx context.Context) (*schema.AudioCodecList, error) {
	// Perform request
	var response schema.AudioCodecList
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("codec")); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
