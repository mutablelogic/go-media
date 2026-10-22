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
	endpoint = "https://api.themoviedb.org/3"

	// rateLimit is the most requests each second, which TMDB allows, and
	// which a quota can't go above
	rateLimit = 40.0

	// DefaultQuota is the most requests each minute, unless OptQuota sets
	// another, which is the rate limit
	DefaultQuota = uint(rateLimit * 60)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a TMDB client with the API Read Access Token, and options. Its
// requests are limited to DefaultQuota each minute, unless OptQuota sets
// another.
func New(apikey string, opts ...client.ClientOpt) (*Client, error) {
	c := new(Client)
	opts = append([]client.ClientOpt{
		client.OptEndpoint(endpoint),
		client.OptReqToken(client.Token{
			Scheme: client.Bearer,
			Value:  apikey,
		}),
		client.OptRateLimit(rateLimit),
	}, opts...)
	cl, err := client.New(opts...)
	if err != nil {
		return nil, err
	} else {
		c.Client = cl
	}
	return c, nil
}

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// OptQuota limits the client's requests to the quota each minute, by spacing
// them evenly. The quota can't be more than TMDB allows, which is
// DefaultQuota, and a quota of zero is DefaultQuota.
func OptQuota(quota uint) client.ClientOpt {
	if quota == 0 || quota > DefaultQuota {
		quota = DefaultQuota
	}
	return client.OptRateLimit(float32(quota) / 60)
}
