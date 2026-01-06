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

// RegisterFormatHandlers registers HTTP handlers for format listing and retrieval
// on the provided router with the given path prefix. The manager must be non-nil.
func RegisterFormatHandlers(router *http.ServeMux, prefix string, manager *task.Manager) {
	// List objects across all databases
	router.HandleFunc(types.JoinPath(prefix, "format"), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = formatList(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func formatList(w http.ResponseWriter, r *http.Request, manager *task.Manager) error {
	// Parse request
	var req schema.ListFormatRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the objects
	response, err := manager.ListFormats(r.Context(), &req)
	if err != nil {
		return httpresponse.Error(w, httperr(err))
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
