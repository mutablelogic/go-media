package googlephotos

import (
	"fmt"
	"net/http"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-media/pkg/googleclient"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SearchOpt func(params *mediaItemSearch)
type DownloadOpt func(params url.Values)
type UploadOpt func(params url.Values)

////////////////////////////////////////////////////////////////////////////////
// CLIENT OPTIONS

func OptPageSize(v uint) googleclient.ClientOpt {
	return func(params url.Values, _ *http.Request) googleclient.ClientOptDone {
		params.Set("pageSize", fmt.Sprint(v))
		return nil
	}
}

func OptPageToken(v *string) googleclient.ClientOpt {
	return func(params url.Values, _ *http.Request) googleclient.ClientOptDone {
		params.Set("pageToken", *v)
		// Return a function that will set the token to the next page
		return func(out interface{}) {
			if arr, ok := out.(*Array); ok {
				*v = arr.NextPageToken
			}
		}
	}
}

func OptExcludeNonAppCreatedData(v bool) googleclient.ClientOpt {
	return func(params url.Values, _ *http.Request) googleclient.ClientOptDone {
		params.Set("excludeNonAppCreatedData", fmt.Sprint(v))
		return nil
	}
}

func OptMimeType(v string) googleclient.ClientOpt {
	return func(params url.Values, req *http.Request) googleclient.ClientOptDone {
		if v != "" {
			req.Header.Set("X-Goog-Upload-Content-Type", v)
		} else {
			req.Header.Del("X-Goog-Upload-Content-Type")
		}
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// SEARCH OPTIONS

func OptAlbumId(v string) SearchOpt {
	return func(params *mediaItemSearch) {
		params.AlbumId = v
	}
}

func OptMediaTypeVideo() SearchOpt {
	return func(params *mediaItemSearch) {
		params.Filters.MediaTypeFilter.MediaTypes = []string{"VIDEO"}
	}
}

func OptMediaTypePhoto() SearchOpt {
	return func(params *mediaItemSearch) {
		params.Filters.MediaTypeFilter.MediaTypes = []string{"PHOTO"}
	}
}

func OptMediaTypeAll() SearchOpt {
	return func(params *mediaItemSearch) {
		params.Filters.MediaTypeFilter.MediaTypes = []string{"ALL_MEDIA"}
	}
}

func OptContentFilterInclude(v ...string) SearchOpt {
	return func(params *mediaItemSearch) {
		params.Filters.ContentFilter.IncludedContentCategories = append(params.Filters.ContentFilter.IncludedContentCategories, v...)
	}
}

func OptContentFilterExclude(v ...string) SearchOpt {
	return func(params *mediaItemSearch) {
		params.Filters.ContentFilter.ExcludedContentCategories = append(params.Filters.ContentFilter.ExcludedContentCategories, v...)
	}
}

func OptFeatureFilter(v ...string) SearchOpt {
	return func(params *mediaItemSearch) {
		params.Filters.FeatureFilter.IncludedFeatures = append(params.Filters.FeatureFilter.IncludedFeatures, v...)
	}
}

////////////////////////////////////////////////////////////////////////////////
// DOWNLOAD OPTIONS

func OptWidthHeight(w, h uint, crop bool) DownloadOpt {
	return func(params url.Values) {
		if w != 0 {
			params.Set("w", fmt.Sprint(w))
		} else {
			params.Del("w")
		}
		if h != 0 {
			params.Set("h", fmt.Sprint(h))
		} else {
			params.Del("h")
		}
		if crop {
			params.Set("c", "")
		} else {
			params.Del("c")
		}
	}
}

func OptMetadata() DownloadOpt {
	return func(params url.Values) {
		params.Set("d", "")
	}
}

func OptVideo(overlay bool) DownloadOpt {
	return func(params url.Values) {
		params.Set("dv", "")
		if !overlay {
			params.Set("no", "")
		} else {
			params.Del("no")
		}
	}
}
