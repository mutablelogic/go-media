package googlephotos

import (
	"encoding/json"
	"path/filepath"

	"github.com/mutablelogic/go-media/pkg/googleclient"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Album struct {
	Id                    string    `json:"id"`
	Title                 string    `json:"title"`
	ProductUrl            string    `json:"productUrl"`
	IsWritable            bool      `json:"isWriteable"`
	ShareInfo             ShareInfo `json:"shareInfo"`
	MediaItemsCount       string    `json:"mediaItemsCount"`
	CoverPhotoBaseUrl     string    `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemId string    `json:"coverPhotoMediaItemId"`
}

type ShareInfo struct {
	SharedAlbumOptions `json:"sharedAlbumOptions"`
	SharableUrl        string `json:"shareableUrl,omitempty"`
	ShareToken         string `json:"shareToken,omitempty"`
	IsJoined           bool   `json:"isJoined"`
	IsOwned            bool   `json:"isOwned"`
	IsJoinable         bool   `json:"isJoinable"`
}

type SharedAlbumOptions struct {
	IsCollaborative bool `json:"isCollaborative"`
	IsCommentable   bool `json:"isCommentable"`
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func AlbumList(client *googleclient.Client, opts ...googleclient.ClientOpt) ([]Album, error) {
	var result Array
	if err := client.Get("/v1/albums", &result, opts...); err != nil {
		return nil, err
	} else {
		return result.Albums, nil
	}
}

func AlbumGet(client *googleclient.Client, id string) (Album, error) {
	var result Album

	path := filepath.Join("/v1/albums", id)
	if err := client.Get(path, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

func AlbumCreate(client *googleclient.Client, id string) (Album, error) {
	var result Album

	path := filepath.Join("/v1/albums", id)
	if err := client.Get(path, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (a Album) String() string {
	b, _ := json.MarshalIndent(a, "", "  ")
	return string(b)
}
