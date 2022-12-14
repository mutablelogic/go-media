package googlephotos

import (
	"encoding/json"

	// Packages
	"github.com/mutablelogic/go-media/pkg/googleclient"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MediaItem struct {
	Id              string `json:"id"`
	Description     string `json:"description"`
	ProductUrl      string `json:"productUrl"`
	BaseUrl         string `json:"baseUrl"`
	MimeType        string `json:"mimeType"`
	MediaMetadata   `json:"mediaMetadata"`
	ContributorInfo `json:"contributorInfo,omitempty"`
	Filename        string `json:"filename"`
}

type MediaMetadata struct {
	CreationTime string `json:"creationTime"`
	Width        string `json:"width"`
	Height       string `json:"height"`
	Photo        `json:"photo,omitempty"`
	Video        `json:"video,omitempty"`
}

type ContributorInfo struct {
	ProfilePictureBaseUrl string `json:"profilePictureBaseUrl,omitempty"`
	DisplayName           string `json:"displayName,omitempty"`
}

type Photo struct {
	CameraMake      string  `json:"cameraMake,omitempty"`
	CameraModel     string  `json:"cameraModel,omitempty"`
	FocalLength     float64 `json:"focalLength,omitempty"`
	ApertureFNumber float64 `json:"apertureFNumber,omitempty"`
	IsoEquivalent   float64 `json:"isoEquivalent,omitempty"`
	ExposureTime    string  `json:"exposureTime,omitempty"`
}

type Video struct {
	CameraMake      string  `json:"cameraMake,omitempty"`
	CameraModel     string  `json:"cameraModel,omitempty"`
	FramesPerSecond float64 `json:"fps,omitempty"`
	Status          string  `json:"status,omitempty"`
}

type mediaItemSearch struct {
	AlbumId   string           `json:"albumId,omitempty"`
	PageSize  uint             `json:"pageSize,omitempty"`
	PageToken string           `json:"pageToken,omitempty"`
	OrderBy   string           `json:"orderBy,omitempty"`
	Filters   mediaItemFilters `json:"filters,omitempty"`
}

type mediaItemFilters struct {
	DateFilter               `json:"dateFilter,omitempty"`
	ContentFilter            `json:"contentFilter,omitempty"`
	MediaTypeFilter          `json:"mediaTypeFilter,omitempty"`
	FeatureFilter            `json:"featureFilter,omitempty"`
	IncludeArchivedMedia     bool `json:"includeArchivedMedia,omitempty"`
	ExcludeNonAppCreatedData bool `json:"excludeNonAppCreatedData,omitempty"`
}

type DateFilter struct{}

type ContentFilter struct {
	IncludedContentCategories []string `json:"includedContentCategories,omitempty"`
	ExcludedContentCategories []string `json:"excludedContentCategories,omitempty"`
}

type MediaTypeFilter struct {
	MediaTypes []string `json:"mediaTypes,omitempty"`
}

type FeatureFilter struct {
	IncludedFeatures []string `json:"includedFeatures,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func MediaItemList(client *googleclient.Client, opts ...googleclient.ClientOpt) ([]*MediaItem, error) {
	var result Array
	if err := client.Get("/v1/mediaItems", &result, opts...); err != nil {
		return nil, err
	} else {
		return result.MediaItems, nil
	}
}

func MediaItemGet(client *googleclient.Client, opts ...googleclient.ClientOpt) (*MediaItem, error) {
	var result MediaItem
	if err := client.Get("/v1/mediaItems", &result, opts...); err != nil {
		return nil, err
	} else {
		return &result, nil
	}
}

func MediaItemSearch(client *googleclient.Client, opts ...SearchOpt) ([]*MediaItem, error) {
	var result Array

	// Set parameters for search
	in := new(mediaItemSearch)
	for _, opt := range opts {
		opt(in)
	}

	// Perform the search
	if err := client.Post("/v1/mediaItems:search", in, &result); err != nil {
		return nil, err
	} else {
		return result.MediaItems, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m MediaItem) String() string {
	b, _ := json.MarshalIndent(m, "", "  ")
	return string(b)
}
