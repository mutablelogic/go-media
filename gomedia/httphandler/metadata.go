package httphandler

import (
	_ "embed"
	"errors"
	"io"
	"net/http"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	task "github.com/mutablelogic/go-media/gomedia/task"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type FormData struct {
	File types.File `json:"file" form:"file" validate:"required"`
}

// namedReader pairs a reader with an explicit name, so a multipart upload's
// filename survives as far as Probe's gomedia.NamedReader check - form.File
// only carries the filename in its own Path field, which isn't a reader.
type namedReader struct {
	io.Reader
	name string
}

func (r namedReader) Name() string { return r.name }

// Seek forwards to the embedded reader if it supports seeking - form.File.Body
// (go-server's httprequest multipart handling) already does, since a
// multipart.FileHeader is opened via the stdlib's ParseMultipartForm, which
// keeps small parts as an in-memory ReadSeeker and spills larger ones to a
// real, seekable temp file. Without this, embedding io.Reader as an
// interface-typed field would silently drop that capability: Go only
// promotes methods declared on a field's static type, not ones the runtime
// value stored in it happens to have.
func (r namedReader) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := r.Reader.(io.Seeker)
	if !ok {
		return 0, errors.New("reader does not support seeking")
	}
	return seeker.Seek(offset, whence)
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

//go:embed README.md
var readme []byte

func RegisterMetadataHandlers(manager *manager.Media, router *httprouter.Router) error {
	// Parse the documentation
	documentation := openapi.ParseMarkdown(readme)

	// Add the documentation to the router
	router.Spec().AddTag("Metadata", documentation.Section(2, "Media Metadata").Body)

	return errors.Join(
		router.Register("probe/media", nil, func(path httprequest.PathItem) {
			path.Tag("Metadata")

			// POST
			path.Post(func(w http.ResponseWriter, r *http.Request) {
				// Format/Opts come from the query string; Reader is filled in
				// below depending on the request's content type.
				var req task.ProbeMediaRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				// multipart/form-data (a single file field named "file") is a
				// special case; any other content type probes the raw request
				// body directly, so a client can just stream media bytes.
				mediaType, _ := types.ParseContentType(r.Header.Get(types.ContentTypeHeader))
				if mediaType == types.ContentTypeFormData {
					var form FormData
					if err := httprequest.Read(r, &form); err != nil {
						httpresponse.Error(w, gomedia.HTTPErr(err))
						return
					}
					defer form.File.Body.Close()
					req.Reader = namedReader{Reader: form.File.Body, name: form.File.Path}
				} else {
					req.Reader = r.Body
				}

				response, err := manager.ProbeMedia(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Probe Media")
				op.Description(documentation.Section(3, "POST /probe/media").Body)
				op.Query(jsonschema.MustFor[task.ProbeMediaRequest]())
				op.RequestBody(jsonschema.MustFor[FormData](), types.ContentTypeFormData)
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[task.ProbeResponse](), "Media Format, Streams, and Metadata")
			})
		}),
		router.Register("probe/source", nil, func(path httprequest.PathItem) {
			path.Tag("Metadata")

			// POST - no request body (Url is a plain string field, so the
			// whole request decodes via httprequest.Query directly, unlike
			// the *url.URL it used to be).
			path.Post(func(w http.ResponseWriter, r *http.Request) {
				var req task.ProbeSourceRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				response, err := manager.ProbeSource(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Probe Source")
				op.Description(documentation.Section(3, "POST /probe/source").Body)
				op.Query(jsonschema.MustFor[task.ProbeSourceRequest]())
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[task.ProbeResponse](), "Media Format, Streams, and Metadata")
			})
		}),
	)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS
