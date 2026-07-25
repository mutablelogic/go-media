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

func (c *Client) ListSampleFormats(ctx context.Context, req schema.SampleFormatListRequest) (*schema.SampleFormatList, error) {
	// Perform request
	var response schema.SampleFormatList
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("sampleformat"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
