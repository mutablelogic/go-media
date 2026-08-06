package httphandler

import (
	"context"
	"encoding/json"
	"net/http"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	taskencoder "github.com/mutablelogic/go-media/task/encoder"
	taskmanager "github.com/mutablelogic/go-media/task/manager"
	taskschema "github.com/mutablelogic/go-media/task/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// EncodeFormData is the multipart/form-data shape POST /encode expects - the
// media to encode as a file part, and the rest of the request (output
// format, audio/video/subtitle profiles) as a JSON-encoded string part,
// since taskencoder.EncodeRequest's profile fields are too complex for the
// generic query/form decoder to fill directly.
type EncodeFormData struct {
	File    types.File `json:"file" form:"file" validate:"required"`
	Request string     `json:"request" form:"request" validate:"required"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// RegisterEncoderHandlers registers the encoder HTTP handlers. It takes both
// the media manager (to start an encode) and the task manager (to stream
// that task's own events for the text/event-stream response) as separate
// arguments, mirroring how they're wired independently in cmd/gomedia/server.go.
func RegisterEncoderHandlers(manager *manager.Media, tasks *taskmanager.Manager, router *httprouter.Router) error {
	// Parse the documentation
	documentation := openapi.ParseMarkdown(readme)

	// Add the documentation to the router
	router.Spec().AddTag("Encoder", documentation.Section(2, "Encoder").Body)

	return router.Register("encode", nil, func(path httprequest.PathItem) {
		path.Tag("Encoder")

		// POST
		path.Post(func(w http.ResponseWriter, r *http.Request) {
			var form EncodeFormData
			if err := httprequest.Read(r, &form); err != nil {
				httpresponse.Error(w, gomedia.HTTPErr(err))
				return
			}

			// The file part supplies the reader; the request part supplies
			// everything else (EncodeRequest.Reader is excluded from JSON, so
			// unmarshaling into it directly is safe).
			var req taskencoder.EncodeRequest
			if err := json.Unmarshal([]byte(form.Request), &req); err != nil {
				form.File.Body.Close()
				httpresponse.Error(w, gomedia.HTTPErr(gomedia.ErrBadParameter.Withf("invalid request: %v", err)))
				return
			}
			req.Reader = namedReader{Reader: form.File.Body, name: form.File.Path}

			status, err := manager.Encode(r.Context(), req)
			if err != nil {
				form.File.Body.Close()
				httpresponse.Error(w, gomedia.HTTPErr(err))
				return
			}

			// The task reads from form.File.Body asynchronously, and keeps
			// doing so after this handler returns - Encode doesn't wait for
			// it to finish. Closing the body on a defer tied to the handler,
			// as every earlier return path in this function does, would
			// race the task's own reads - closing, and for a large upload
			// spooled to disk deleting, the file out from under it. Hand
			// ownership off to the task's own lifetime instead: a
			// background wait, using its own context since it must outlive
			// this request.
			go func() {
				_, _ = tasks.Wait(context.Background(), status.UUID)
				form.File.Body.Close()
			}()

			accept, err := types.AcceptContentType(r)
			if err != nil {
				httpresponse.Error(w, gomedia.HTTPErr(gomedia.ErrBadParameter.Withf("invalid accept header: %v", err)))
				return
			}

			if accept != types.ContentTypeTextStream {
				httpresponse.JSON(w, http.StatusAccepted, httprequest.Indent(r), status)
				return
			}

			stream := httpresponse.NewTextStream(w)
			if stream == nil {
				httpresponse.Error(w, gomedia.HTTPErr(gomedia.ErrInternalError.With("failed to open event stream")))
				return
			}
			defer stream.Close()

			// Stop once this task's own finished event arrives, rather than
			// waiting for the client to disconnect or the task manager itself
			// to stop.
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			_ = tasks.Subscribe(ctx, func(e *taskschema.Event) {
				if e.Status.UUID != status.UUID {
					return
				}
				stream.Write(string(e.Name), e.Status)
				if e.Name == taskschema.EventFinished {
					cancel()
				}
			})
		}, func(op httprequest.PathOperation) {
			op.Summary("Encode Media")
			op.Description(documentation.Section(3, "POST /encode").Body)
			op.RequestBody(jsonschema.MustFor[EncodeFormData](), types.ContentTypeFormData)
			op.JSONResponse(http.StatusAccepted, jsonschema.MustFor[taskschema.Status](), "Task Status")
			op.Response(http.StatusOK, types.ContentTypeTextStream, "Stream of Task Events")
			op.ErrorResponse(http.StatusBadRequest)
		})
	})
}
