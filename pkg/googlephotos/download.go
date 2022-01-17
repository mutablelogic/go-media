package googlephotos

import (
	"errors"
	"fmt"
	"io"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-media/pkg/googleclient"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultBufferSize = 1024
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Download a media item from Google Photos, given the item and any download options. If no options
// are provided
func DownloadMediaItem(client *googleclient.Client, w io.Writer, item *MediaItem, opts ...DownloadOpt) error {
	if client == nil || w == nil || item == nil || item.BaseUrl == "" {
		return ErrBadParameter.With("DownloadMediaItem")
	}

	// Download
	return downloadUrl(client, w, item.BaseUrl, opts...)
}

// Download a media item profile picture from Google Photos, given the item and any download options. If no options
// are provided
func DownloadMediaItemProfilePicture(client *googleclient.Client, w io.Writer, item *MediaItem, opts ...DownloadOpt) error {
	if client == nil || w == nil || item == nil || item.ProfilePictureBaseUrl == "" {
		return ErrBadParameter.With("DownloadMediaItemProfilePicture")
	}

	// Download
	return downloadUrl(client, w, item.ProfilePictureBaseUrl, opts...)
}

// Download a media item from Google Photos, given the item and any download options. If no options
// are provided
func DownloadAlbumCover(client *googleclient.Client, w io.Writer, item *Album, opts ...DownloadOpt) error {
	if client == nil || w == nil || item == nil || item.CoverPhotoBaseUrl == "" {
		return ErrBadParameter.With("DownloadAlbumCover")
	}

	// Download
	return downloadUrl(client, w, item.CoverPhotoBaseUrl, opts...)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func downloadUrl(client *googleclient.Client, w io.Writer, u string, opts ...DownloadOpt) error {
	// Make a url
	u_, err := url.Parse(u)
	if err != nil {
		return err
	}

	// Get parameters
	params := url.Values{}
	for _, opt := range opts {
		opt(params)
	}

	// Append parameters onto the path
	sep := "="
	for key := range params {
		u_.Path = u_.Path + sep + key + params[key][0]
		sep = "-"
	}

	fmt.Println(u_.String())

	// Get bytes
	resp, err := client.Client.Get(u_.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for response errors
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return googleclient.NewError(resp)
	}

	// Read until EOF
	data := make([]byte, defaultBufferSize)
	for {
		if n, err := resp.Body.Read(data); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		} else if _, err := w.Write(data[:n]); err != nil {
			return err
		}
	}

	// Return success
	return nil
}
