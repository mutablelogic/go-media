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

func (c *Client) CreateAudioProfile(ctx context.Context, req schema.AudioProfileMeta) (*schema.AudioProfile, error) {
	r, err := client.NewJSONRequest(req)
	if err != nil {
		return nil, err
	}
	// Perform request
	var response schema.AudioProfile
	if err := c.DoWithContext(ctx, r, &response, client.OptPath("audio")); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
