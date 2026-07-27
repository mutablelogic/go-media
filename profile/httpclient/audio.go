package httpclient

import (
	"context"
	"net/http"

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

func (c *Client) GetAudioProfile(ctx context.Context, uuid string) (*schema.AudioProfile, error) {
	// Perform request
	var response schema.AudioProfile
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("audio", uuid)); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}

func (c *Client) DeleteAudioProfile(ctx context.Context, uuid string) (*schema.AudioProfile, error) {
	// Perform request
	var response schema.AudioProfile
	if err := c.DoWithContext(ctx, client.MethodDelete, &response, client.OptPath("audio", uuid)); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}

func (c *Client) UpdateAudioProfile(ctx context.Context, uuid string, req schema.AudioProfileMeta) (*schema.AudioProfile, error) {
	r, err := client.NewJSONRequestEx(http.MethodPatch, req, types.ContentTypeAny)
	if err != nil {
		return nil, err
	}

	// Perform request
	var response schema.AudioProfile
	if err := c.DoWithContext(ctx, r, &response, client.OptPath("audio", uuid)); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
