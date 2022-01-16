package googlephotos

import (
	"fmt"
	"net/url"

	// Packages
	"github.com/mutablelogic/go-media/pkg/googleclient"
)

func OptPageSize(v uint) googleclient.ClientOpt {
	return func(params url.Values) {
		params.Set("pageSize", fmt.Sprint(v))
	}
}

func OptPageToken(v string) googleclient.ClientOpt {
	return func(params url.Values) {
		params.Set("pageToken", v)
	}
}

func OptExcludeNonAppCreatedData(v bool) googleclient.ClientOpt {
	return func(params url.Values) {
		params.Set("excludeNonAppCreatedData", fmt.Sprint(v))
	}
}
