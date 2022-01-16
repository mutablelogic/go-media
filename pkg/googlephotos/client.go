package googlephotos

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

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Array struct {
	Albums        []Album     `json:"albums,omitempty"`
	SharedAlbums  []Album     `json:"sharedAlbums,omitempty"`
	MediaItems    []MediaItem `json:"mediaItems,omitempty"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}
