package chromaprint

import (
	"fmt"
	"net/url"
	"time"

	// Packages
	"github.com/mutablelogic/go-client"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
	key string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Register a clientId: https://acoustid.org/login
	defaultApiKey = "-YectPtndAM"

	// The API endpoint
	endPoint = "https://api.acoustid.org/v2"

	// defaultQps rate limits number of requests per second
	defaultQps = 3
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new client
func NewClient(ApiKey string, opts ...client.ClientOpt) (*Client, error) {
	// Check for missing API key
	if ApiKey == "" {
		ApiKey = defaultApiKey
	}

	// Create client
	opts = append(opts, client.OptEndpoint(endPoint))
	client, err := client.New(opts...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{
		Client: client,
		key:    ApiKey,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// LOOKUP

// Lookup a fingerprint with a duration and the metadata that needs to be
// returned
func (c *Client) Lookup(fingerprint string, duration time.Duration, flags Meta) ([]*ResponseMatch, error) {
	// Check incoming parameters
	if fingerprint == "" || duration == 0 || flags == META_NONE {
		return nil, ErrBadParameter.With("Lookup")
	}

	// Set URL parameters
	params := url.Values{}
	params.Set("client", c.key)
	params.Set("fingerprint", fingerprint)
	params.Set("duration", fmt.Sprint(duration.Truncate(time.Second).Seconds()))
	params.Set("meta", flags.String())

	// Request -> Response
	var response Response
	if err := c.Do(nil, &response, client.OptPath("lookup"), client.OptQuery(params)); err != nil {
		return nil, err
	} else {
		return response.Results, nil
	}
}
