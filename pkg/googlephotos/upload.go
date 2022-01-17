package googlephotos

import (
	"io"
	"net/http"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-media/pkg/googleclient"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Upload a media item to Google Photos, given a byte stream. Returns an upload
// token which can be used to create a media item
// https://developers.google.com/photos/library/guides/upload-media#creating-media-item
func UploadMediaItem(client *googleclient.Client, r io.Reader, opts ...googleclient.ClientOpt) (string, error) {
	var result string
	if client == nil || r == nil {
		return "", ErrBadParameter.With("UploadMediaItem")
	}
	opts = append(opts, func(params url.Values, req *http.Request) {
		req.Header.Set("X-Goog-Upload-Protocol", "raw")
	})
	if err := client.PostBinary("/v1/uploads", r, &result, opts...); err != nil {
		return "", err
	} else {
		return result, nil
	}
}
