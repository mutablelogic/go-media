package httpclient

import (
	// Packages
	client "github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Client is a profile HTTP client that wraps the base HTTP client
type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new profile HTTP client with the given base URL and options.
// The url parameter should point to the profile API endpoint, e.g.
// "http://localhost:8080/api".
func New(url string, opts ...client.ClientOpt) (*Client, error) {
	c := new(Client)
	cl, err := client.New(append(opts, client.OptEndpoint(url))...)
	if err != nil {
		return nil, err
	}
	c.Client = cl
	return c, nil
}
