package httphandler

import (
	"net/http"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterPixelFormatHandlers registers HTTP handlers for pixel format listing and retrieval
// on the provided router with the given path prefix. The manager must be non-nil.
func RegisterPixelFormatHandlers(router *http.ServeMux, prefix string, manager *task.Manager) {
	// List objects across all databases
	router.HandleFunc(types.JoinPath(prefix, "pixelformat"), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = pixelFormatList(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func pixelFormatList(w http.ResponseWriter, r *http.Request, manager *task.Manager) error {
	// Parse request
	var req schema.ListPixelFormatRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the objects
	response, err := manager.ListPixelFormats(r.Context(), &req)
	if err != nil {
		return httpresponse.Error(w, httperr(err))
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
