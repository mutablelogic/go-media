package googleclient

import (
	"net/http"
	"net/url"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientOpt func(params url.Values, req *http.Request)

////////////////////////////////////////////////////////////////////////////////
// CLIENT OPTIONS

func OptHeader(key, value string) ClientOpt {
	return func(params url.Values, req *http.Request) {
		req.Header.Set(key, value)
	}
}
