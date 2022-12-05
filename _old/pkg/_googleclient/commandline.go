package googleclient

import (
	"context"

	// Packages
	"golang.org/x/oauth2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type CommandLineAuth struct {
	VerificationURL string `json:"verification_url"`
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Start device authentication
func (client *Client) CommandLineAuth() (*CommandLineAuth, error) {
	var result CommandLineAuth

	result.VerificationURL = client.Config.AuthCodeURL(client.Name)

	return &result, nil
}

// Start device authentication
func (client *Client) CommandLineToken(ctx context.Context, code string) (*oauth2.Token, error) {
	// Handle the exchange code to initiate a transport.
	token, err := client.Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return token, nil
}
