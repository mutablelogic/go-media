package metadata

import (
	"slices"
	"time"

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

	// The content type of the stream, which GetMetadata sets
	contentType string

	// The duration of the stream, when the caller knows it
	duration time.Duration

	// The year of the stream, such as its release year, when the caller
	// knows it, or zero
	year int
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

// IsText reports whether the stream is a text document with a type which
// isn't a text type, such as a playlist with an audio type. Media handlers,
// such as ffmpeg's, don't read it, as it isn't media.
func (o *Opts) IsText() bool {
	return slices.Contains(textTypes, o.contentType)
}

// Duration returns the duration of the stream set via WithDuration, or zero
// when it isn't known
func (o *Opts) Duration() time.Duration {
	return o.duration
}

// Year returns the year of the stream set via WithYear, or zero when it
// isn't known
func (o *Opts) Year() int {
	return o.year
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

// WithDuration sets the duration of the stream, when the caller knows it,
// such as from the stream's own metadata. The TMDB handler then only matches
// a movie whose runtime is within 10% of it.
func WithDuration(d time.Duration) Option {
	return func(o *Opts) error {
		if d < 0 {
			return gomedia.ErrBadParameter.Withf("invalid duration %v", d)
		}
		o.duration = d
		return nil
	}
}

// WithYear sets the year of the stream, such as its release year, when the
// caller knows it, such as from the stream's own metadata. The TMDB handler
// then matches a movie whose title is the stream's, and whose release year
// is the year, whatever its runtime. A year in the stream's name is used
// rather than it. Zero is no year.
func WithYear(year int) Option {
	return func(o *Opts) error {
		if year < 0 || year > 9999 {
			return gomedia.ErrBadParameter.Withf("invalid year %d", year)
		}
		o.year = year
		return nil
	}
}
