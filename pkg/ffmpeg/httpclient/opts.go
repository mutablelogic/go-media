package httpclient

import (
	"net/url"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	url.Values
}

// Opt is an option to set on the client request.
type Opt func(*opt) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opts ...Opt) (*opt, error) {
	o := new(opt)
	o.Values = make(url.Values)
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}
