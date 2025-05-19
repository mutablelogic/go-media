package chromaprint

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"
	"unsafe"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/segmenter"
	"github.com/mutablelogic/go-media/sys/chromaprint"

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

////////////////////////////////////////////////////////////////////////////////
// FINGERPRINT

// Match a media file and lookup any matches
func (c *Client) Match(ctx context.Context, r io.Reader, flags Meta) ([]*ResponseMatch, error) {
	// Create a segmenter
	segmenter, err := segmenter.NewReader(r, 0, 32000)
	if err != nil {
		return nil, err
	}
	defer segmenter.Close()

	// Create a fingerprinting context
	fp := chromaprint.NewChromaprint(chromaprint.ALGORITHM_DEFAULT)
	if fp == nil {
		return nil, media.ErrInternalError.With("chromaprint.NewChromaprint")
	}
	defer fp.Free()

	// Start the fingerprinting
	if err := fp.Start(32000, 1); err != nil {
		return nil, err
	}

	// Perform fingerprinting
	if err := segmenter.DecodeInt16(ctx, func(timestamp time.Duration, data []int16) error {
		return fp.WritePtr(uintptr(unsafe.Pointer(&data[0])), len(data))
	}); err != nil {
		return nil, err
	}

	// Complete fingerprinting
	if err := fp.Finish(); err != nil {
		return nil, err
	}

	value, err := fp.GetFingerprint()
	if err != nil {
		return nil, err
	}

	// Lookup fingerprint
	return c.Lookup(value, segmenter.Duration(), flags)
}
