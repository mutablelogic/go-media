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

	// sample rate used for fingerprinting
	sampleRate = 32000

	// maxFingerprintDuration is the maximum duration to fingerprint
	// Chromaprint only needs ~120 seconds for a reliable fingerprint
	maxFingerprintDuration = 120 * time.Second
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new client with rate limiting (3 requests per second by default)
func NewClient(ApiKey string, opts ...client.ClientOpt) (*Client, error) {
	// Check for missing API key
	if ApiKey == "" {
		ApiKey = defaultApiKey
	}

	// Create client with rate limiting and endpoint
	opts = append(opts,
		client.OptEndpoint(endPoint),
		client.OptRateLimit(defaultQps),
	)
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
	}

	// Check for API error
	if response.Status != "ok" {
		if response.Error.Message != "" {
			return nil, ErrBadParameter.Withf("acoustid: %s", response.Error.Message)
		}
		return nil, ErrBadParameter.With("acoustid: unknown error")
	}

	return response.Results, nil
}

////////////////////////////////////////////////////////////////////////////////
// FINGERPRINT

// FingerprintResult contains the fingerprint and duration of the audio
type FingerprintResult struct {
	Fingerprint string
	Duration    time.Duration
}

// Fingerprint generates an audio fingerprint from the reader. The "dur" parameter
// specifies the full track duration. Only the first 120 seconds (or less) will be
// fingerprinted, but the returned FingerprintResult.Duration will reflect the full
// track duration. If dur is zero, the duration will be auto-detected from the file.
func Fingerprint(ctx context.Context, r io.Reader, dur time.Duration, opts ...segmenter.Opt) (*FingerprintResult, error) {
	// Limit fingerprinting to maximum needed for reliable matching
	fingerprintDur := dur
	if fingerprintDur <= 0 || fingerprintDur > maxFingerprintDuration {
		fingerprintDur = maxFingerprintDuration
	}

	// Always set segment size, allow user to add more options
	segmenterOpts := append([]segmenter.Opt{segmenter.WithSegmentSize(time.Second)}, opts...)
	seg, err := segmenter.NewFromReader(r, sampleRate, segmenterOpts...)
	if err != nil {
		return nil, err
	}
	defer seg.Close()

	// Create a fingerprinting context
	fp := chromaprint.NewChromaprint(chromaprint.ALGORITHM_DEFAULT)
	if fp == nil {
		return nil, media.ErrInternalError.With("chromaprint.NewChromaprint")
	}
	defer fp.Free()

	// Start the fingerprinting
	if err := fp.Start(sampleRate, 1); err != nil {
		return nil, err
	}

	// Track processed duration
	var processedDuration time.Duration

	// Perform fingerprinting until we reach the fingerprint duration limit
	if err := seg.DecodeInt16(ctx, func(timestamp time.Duration, data []int16) error {
		if timestamp >= fingerprintDur {
			// Stop early - we have enough samples
			return io.EOF
		}

		if err := fp.WritePtr(uintptr(unsafe.Pointer(&data[0])), len(data)); err != nil {
			return err
		}

		// Update processed duration
		sampleDuration := time.Duration(len(data)) * time.Second / time.Duration(sampleRate)
		processedDuration = timestamp + sampleDuration
		return nil
	}); err != nil {
		return nil, err
	}

	// Complete fingerprinting
	if err := fp.Finish(); err != nil {
		return nil, err
	}

	// Get fingerprint value
	value, err := fp.GetFingerprint()
	if err != nil {
		return nil, err
	}

	// Determine final duration to report
	var finalDuration time.Duration
	if dur > 0 {
		// Use explicitly provided duration
		finalDuration = dur
	} else {
		// Auto-detect: prefer full file duration, fall back to processed duration
		finalDuration = seg.Duration()
		if finalDuration <= 0 {
			finalDuration = processedDuration
		}
	}

	return &FingerprintResult{
		Fingerprint: value,
		Duration:    finalDuration,
	}, nil
}

// Match generates a fingerprint from the reader and looks up any matches.
// The "dur" parameter specifies the full track duration. Only the first 120 seconds
// (or less) will be fingerprinted for matching, but "dur" will be reported to AcoustID.
// If dur is zero, the duration will be auto-detected from the file.
func (c *Client) Match(ctx context.Context, r io.Reader, dur time.Duration, flags Meta, opts ...segmenter.Opt) ([]*ResponseMatch, error) {
	// Generate fingerprint
	result, err := Fingerprint(ctx, r, dur, opts...)
	if err != nil {
		return nil, err
	}

	// Lookup fingerprint
	return c.Lookup(result.Fingerprint, result.Duration, flags)
}
