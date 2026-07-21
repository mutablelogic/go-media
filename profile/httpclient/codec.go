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

func (c *Client) ListCodecs(ctx context.Context, req schema.CodecListRequest) (*schema.CodecList, error) {
	// Perform request
	var response schema.CodecList
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("codec"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}

func (c *Client) GetCodec(ctx context.Context, name string) (*schema.Codec, error) {
	// Perform request
	var response schema.Codec
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("codec", name)); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
