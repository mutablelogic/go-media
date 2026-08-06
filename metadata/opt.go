package metadata

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	gomedia "github.com/mutablelogic/go-media"
	tmdb "github.com/mutablelogic/go-media/tmdb/httpclient"
	types "github.com/mutablelogic/go-server/pkg/types"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Opts struct {
	// List of namespaces that the caller is interested in.
	namespaces []string

	// Add OTEL tracing
	tracer trace.Tracer

	// When non-nil, this is the TMDB client used to fetch metadata from TMDB.
	tmdb *tmdb.Client
}

type Option func(*Opts) error

////////////////////////////////////////////////////////////////////////////////
// METHODS

// HasNamespace reports whether ns was requested via WithNamespace,
// case-insensitively. Handlers that do expensive work for a single
// namespace (e.g. artwork extraction) use this to opt in only when that
// namespace was explicitly requested, rather than whenever no namespace
// filter was given at all.
func (o *Opts) HasNamespace(ns string) bool {
	return containsFold(o.namespaces, ns)
}

// TMDB returns the TMDB client set via WithTMDB, or nil if none was set.
// Handlers that need to query TMDB use this to opt out (return nil, nil)
// when no client is configured, rather than erroring.
func (o *Opts) TMDB() *tmdb.Client {
	return o.tmdb
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOptions(opts ...Option) (*Opts, error) {
	o := new(Opts)
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set the otel tracer to use for metadata operations.
func WithTracer(tracer trace.Tracer) Option {
	return func(o *Opts) error {
		o.tracer = tracer
		return nil
	}
}

// Return only the metadata from the given namespaces.
func WithNamespace(ns ...string) Option {
	return func(o *Opts) error {
		for _, n := range ns {
			if !types.IsIdentifier(n) {
				return gomedia.ErrBadParameter.Withf("invalid namespace: %q", n)
			}
		}
		o.namespaces = ns
		return nil
	}
}

// Use The Movie Database (TMDB) client to fetch metadata from TMDB.
func WithTMDB(token string, clientopts ...client.ClientOpt) Option {
	return func(o *Opts) error {
		if client, err := tmdb.New(token, clientopts...); err != nil {
			return err
		} else {
			o.tmdb = client
		}
		return nil
	}
}
