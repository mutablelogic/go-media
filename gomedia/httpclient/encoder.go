package httpclient

import (
	"context"
	"encoding/json"
	"io"

	// Packages
	client "github.com/mutablelogic/go-client"
	multipart "github.com/mutablelogic/go-client/pkg/multipart"
	gomedia "github.com/mutablelogic/go-media"
	taskencoder "github.com/mutablelogic/go-media/task/encoder"
	taskschema "github.com/mutablelogic/go-media/task/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// EncodeEventFunc is called for every task event received when Encode is
// given a non-nil fn. Return io.EOF to stop the stream cleanly, or any other
// error to stop it and have Encode return that error.
type EncodeEventFunc func(*taskschema.Event) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Encode uploads req.Reader for encoding. If fn is nil, it returns as soon
// as the task is started, with its current status - encoding is
// long-running, so this doesn't wait for it to finish; poll or watch the
// task afterwards by its UUID (via the task API). If fn is non-nil, Encode
// instead waits, streaming the task's own events to fn until it finishes,
// ctx is done, or fn itself stops it (io.EOF to stop cleanly, any other
// error to stop and have Encode return that error) - either way, the
// returned status is the last one observed.
//
// req.Output/Audio/Video/Subtitle must already be fully resolved (e.g. via
// profile.NewAudioProfile), same as when calling the task directly - they're
// sent as JSON alongside the uploaded file. If onRead is non-nil, it's
// called with the cumulative number of bytes read from req.Reader as the
// upload progresses.
func (c *Client) Encode(ctx context.Context, req taskencoder.EncodeRequest, onRead func(n int64), fn EncodeEventFunc) (*taskschema.Status, error) {
	if fn == nil {
		body, err := encodeRequestPayload(req, onRead, types.ContentTypeJSON)
		if err != nil {
			return nil, err
		}

		var response taskschema.Status
		if err := c.DoWithContext(ctx, body, &response, client.OptPath("encode")); err != nil {
			return nil, err
		}

		return &response, nil
	}

	body, err := encodeRequestPayload(req, onRead, types.ContentTypeTextStream)
	if err != nil {
		return nil, err
	}

	var last *taskschema.Status
	err = c.DoWithContext(ctx, body, nil,
		client.OptPath("encode"),
		client.OptTextStreamCallback(func(evt client.TextStreamEvent) error {
			// Keep-alive pings carry no data - nothing to decode or deliver.
			if evt.Data == "" {
				return nil
			}
			var status taskschema.Status
			if err := evt.Json(&status); err != nil {
				return err
			}
			last = &status
			return fn(&taskschema.Event{Name: taskschema.EventName(evt.Event), Status: status})
		}),
	)
	if err != nil {
		return last, err
	}

	return last, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// encodeRequestPayload builds the multipart/form-data body POST /encode
// expects - a "file" part streamed from req.Reader, and a "request" part
// holding the rest of req (Output/Audio/Video/Subtitle) as JSON, matching
// gomedia/httphandler's EncodeFormData. accept selects the response mode
// (application/json for Encode, text/event-stream for SubscribeEncode).
func encodeRequestPayload(req taskencoder.EncodeRequest, onRead func(n int64), accept string) (client.Payload, error) {
	if req.Reader == nil {
		return nil, gomedia.ErrBadParameter.With("nil reader")
	}

	name := ""
	if named, ok := req.Reader.(gomedia.NamedReader); ok && named != nil {
		name = named.Name()
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return client.NewStreamingMultipartRequest(struct {
		File    multipart.File `json:"file"`
		Request string         `json:"request"`
	}{
		File: multipart.File{
			Path: name,
			Body: io.NopCloser(NewReader(req.Reader, onRead)),
		},
		Request: string(data),
	}, accept)
}
