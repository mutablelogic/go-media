package httpclient

import (
	"context"
	"io"

	// Packages
	client "github.com/mutablelogic/go-client"
	multipart "github.com/mutablelogic/go-client/pkg/multipart"
	gomedia "github.com/mutablelogic/go-media"
	task "github.com/mutablelogic/go-media/gomedia/task"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Probe uploads req.Reader to the probe endpoint and returns its container
// format, streams, and metadata. req.Format/req.Opts are attached to the
// request as query parameters.
//
// If contentType is multipart/form-data, req.Reader is streamed as a single
// "file" field using go-client's streaming multipart encoder (so the whole
// body is never buffered in memory, matching the server's expected upload
// shape); for any other content type, req.Reader is streamed directly as
// the raw request body with that Content-Type. Either way, the caller
// retains ownership of req.Reader - Probe doesn't close it.
//
// If req.Reader implements gomedia.NamedReader, its name is used as the
// uploaded filename. If onRead is non-nil, it's called with the cumulative
// number of bytes read from req.Reader as the upload progresses, regardless
// of which of the two upload paths is taken.
func (c *Client) Probe(ctx context.Context, req task.ProbeRequest, contentType string, onRead func(n int64)) (*task.ProbeResponse, error) {
	var body client.Payload

	// Determine name of the uploaded file, if any, from the reader.
	name := ""
	if named, ok := req.Reader.(gomedia.NamedReader); ok && named != nil {
		name = named.Name()
	}

	reader := NewReader(req.Reader, onRead)
	if contentType == types.ContentTypeFormData {
		p, err := client.NewStreamingMultipartRequest(struct {
			File multipart.File `json:"file"`
		}{
			File: multipart.File{
				Path: name,
				Body: io.NopCloser(reader),
			},
		}, types.ContentTypeJSON)
		if err != nil {
			return nil, err
		}
		body = p
	} else {
		body = NewPayload(reader, contentType)
	}

	var response task.ProbeResponse
	if err := c.DoWithContext(ctx, body, &response, client.OptPath("probe"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}

	return &response, nil
}
