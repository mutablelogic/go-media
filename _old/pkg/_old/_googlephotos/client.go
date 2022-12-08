package googlephotos

import (
	"context"

	"github.com/mutablelogic/go-media/pkg/googleclient"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Default Client Endpoint
	Endpoint = "https://photoslibrary.googleapis.com"
)

const (
	// Scopes - https://developers.google.com/photos/library/guides/authorization
	Scope      = "https://www.googleapis.com/auth/photoslibrary"
	ScopeRead  = "https://www.googleapis.com/auth/photoslibrary.readonly"
	ScopeWrite = "https://www.googleapis.com/auth/photoslibrary.appendonly"
	ScopeShare = "https://www.googleapis.com/auth/photoslibrary.sharing"
)

const (
	MaxAlbumsPerPage     = 50
	MaxMediaItemsPerPage = 100
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Array struct {
	Albums        []*Album     `json:"albums,omitempty"`
	SharedAlbums  []*Album     `json:"sharedAlbums,omitempty"`
	MediaItems    []*MediaItem `json:"mediaItems,omitempty"`
	NextPageToken string       `json:"nextPageToken,omitempty"`
}

type Client struct {
	*googleclient.Client
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewClient(client *googleclient.Client) *Client {
	if client == nil {
		return nil
	}
	return &Client{client}
}

////////////////////////////////////////////////////////////////////////////////
// CLIENT METHODS

func (c *Client) AlbumList(ctx context.Context, maxAlbums uint, excludeNonAppCreatedData bool) ([]*Album, error) {
	var results []*Album
	var token string

FOR_LOOP:
	for maxAlbums == 0 || maxAlbums > uint(len(results)) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			pageSize := minUint(maxAlbums, MaxAlbumsPerPage)
			if pageSize == 0 {
				pageSize = MaxAlbumsPerPage
			}
			if album, err := AlbumList(c.Client, OptPageToken(&token), OptPageSize(pageSize), OptExcludeNonAppCreatedData(excludeNonAppCreatedData)); err != nil {
				return nil, err
			} else {
				results = append(results, album...)
			}
			if token == "" {
				break FOR_LOOP
			}
		}
	}
	return results, nil
}

func (c *Client) SharedAlbumList(ctx context.Context, maxAlbums uint, excludeNonAppCreatedData bool) ([]*Album, error) {
	var results []*Album
	var token string

FOR_LOOP:
	for maxAlbums == 0 || maxAlbums > uint(len(results)) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			pageSize := minUint(maxAlbums, MaxAlbumsPerPage)
			if pageSize == 0 {
				pageSize = MaxAlbumsPerPage
			}
			if album, err := SharedAlbumList(c.Client, OptPageToken(&token), OptPageSize(pageSize), OptExcludeNonAppCreatedData(excludeNonAppCreatedData)); err != nil {
				return nil, err
			} else {
				results = append(results, album...)
			}
			if token == "" {
				break FOR_LOOP
			}
		}
	}
	return results, nil
}

func (c *Client) MediaList(ctx context.Context, max uint) ([]*MediaItem, error) {
	var results []*MediaItem
	var token string

FOR_LOOP:
	for max == 0 || max > uint(len(results)) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			pageSize := minUint(max, MaxMediaItemsPerPage)
			if pageSize == 0 {
				pageSize = MaxAlbumsPerPage
			}
			if media, err := MediaItemList(c.Client, OptPageToken(&token), OptPageSize(pageSize)); err != nil {
				return nil, err
			} else {
				results = append(results, media...)
			}
			if token == "" {
				break FOR_LOOP
			}
		}
	}
	return results, nil
}
