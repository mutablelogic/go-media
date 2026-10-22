package tmdb

import (
	"bytes"
	"context"
	"strings"

	// Packages
	client "github.com/mutablelogic/go-client"
	transport "github.com/mutablelogic/go-client/pkg/transport"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// imageEndpoint is where TMDB's images are, which is the secure base URL of
// its configuration
const imageEndpoint = "https://image.tmdb.org/t/p"

// The sizes of images, which are the widths TMDB has, or the original
const (
	SizePoster   = "w780"
	SizeBackdrop = "w780"
	SizeStill    = "w300"
	SizeOriginal = "original"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Image returns the image at the path, which is a movie's or show's poster
// or backdrop path, or an episode's still path, such as
// /qJ2tW6WMUDux911r6m7haRef0WH.jpg, in the size, such as SizePoster. The
// image server doesn't need the API token, so it isn't sent, but the request
// counts towards the client's quota.
func (c *Client) Image(ctx context.Context, size, path string) ([]byte, error) {
	if size = strings.TrimSpace(size); size == "" || strings.Contains(size, "/") {
		return nil, httpresponse.ErrBadRequest.Withf("invalid image size %q", size)
	} else if path = strings.TrimSpace(path); path == "" || !strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return nil, httpresponse.ErrBadRequest.Withf("invalid image path %q", path)
	}
	var buf bytes.Buffer
	if err := c.DoWithContext(transport.WithSkipTokenInjection(ctx), nil, &buf,
		client.OptReqEndpoint(imageEndpoint),
		client.OptPath(size, strings.TrimPrefix(path, "/")),
	); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
