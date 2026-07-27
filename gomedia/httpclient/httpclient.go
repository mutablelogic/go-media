package httpclient

import (
	"io"
	"net/http"

	// Packages
	client "github.com/mutablelogic/go-client"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Client is a profile HTTP client that wraps the base HTTP client
type Client struct {
	*client.Client
}

// reader wraps an io.Reader, invoking onRead (if non-nil) with the
// cumulative number of bytes read after each Read call. Used to report
// upload progress, for both a raw payload and a reader embedded in a
// multipart file field.
type reader struct {
	io.Reader
	total  int64
	onRead func(total int64)
}

var _ io.Reader = (*reader)(nil)

// payload streams a reader directly as the request body with a given
// Content-Type, for uploads that aren't multipart/form-data.
type payload struct {
	io.Reader
	contentType string
}

var _ client.Payload = (*payload)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new profile HTTP client with the given base URL and options.
// The url parameter should point to the profile API endpoint, e.g.
// "http://localhost:8080/api".
func New(url string, opts ...client.ClientOpt) (*Client, error) {
	c := new(Client)
	cl, err := client.New(append(opts, client.OptEndpoint(url))...)
	if err != nil {
		return nil, err
	}
	c.Client = cl
	return c, nil
}

func NewReader(r io.Reader, onRead func(total int64)) *reader {
	return &reader{Reader: r, onRead: onRead}
}

func NewPayload(r io.Reader, contentType string) *payload {
	return &payload{Reader: r, contentType: contentType}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *reader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 {
		r.total += int64(n)
		if r.onRead != nil {
			r.onRead(r.total)
		}
	}
	return n, err
}

func (p *payload) Method() string {
	return http.MethodPost
}

func (p *payload) Accept() string {
	return types.ContentTypeJSON
}

func (p *payload) Type() string {
	return p.contentType
}
