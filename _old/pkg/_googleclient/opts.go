package googleclient

import (
	"net/http"
	"net/url"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientOptDone func(interface{})
type ClientOpt func(url.Values, *http.Request) ClientOptDone

////////////////////////////////////////////////////////////////////////////////
// CLIENT OPTIONS

func OptHeader(key, value string) ClientOpt {
	return func(params url.Values, req *http.Request) ClientOptDone {
		req.Header.Set(key, value)
		return nil
	}
}
