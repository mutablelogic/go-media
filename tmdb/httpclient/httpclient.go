// Package tmdb implements a client for the TMDB API
// https://developer.themoviedb.org/docs/getting-started
package tmdb

import (
	client "github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Client is a profile HTTP client that wraps the base HTTP client
type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	endpoint  = "https://api.themoviedb.org/3"
	rateLimit = 40.0 // 40 requests per second
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new profile HTTP client with the given base URL and options.
// The url parameter should point to the profile API endpoint, e.g.
// "http://localhost:8080/api".
func New(apikey string, opts ...client.ClientOpt) (*Client, error) {
	c := new(Client)
	opts = append(opts, []client.ClientOpt{
		client.OptEndpoint(endpoint),
		client.OptReqToken(client.Token{
			Scheme: client.Bearer,
			Value:  apikey,
		}),
		client.OptRateLimit(rateLimit),
	}...)
	cl, err := client.New(opts...)
	if err != nil {
		return nil, err
	} else {
		c.Client = cl
	}
	return c, nil
}
