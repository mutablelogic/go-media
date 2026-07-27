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

func (c *Client) ListPixelFormats(ctx context.Context, req schema.PixelFormatListRequest) (*schema.PixelFormatList, error) {
	// Perform request
	var response schema.PixelFormatList
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("pixelformat"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
