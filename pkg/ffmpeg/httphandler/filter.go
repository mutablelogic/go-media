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

// RegisterFilterHandlers registers HTTP handlers for filter listing and retrieval
// on the provided router with the given path prefix. The manager must be non-nil.
func RegisterFilterHandlers(router *http.ServeMux, prefix string, manager *task.Manager) {
	// List available filters
	router.HandleFunc(types.JoinPath(prefix, "filter"), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = filterList(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func filterList(w http.ResponseWriter, r *http.Request, manager *task.Manager) error {
	// Parse request
	var req schema.ListFilterRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the objects
	response, err := manager.ListFilters(r.Context(), &req)
	if err != nil {
		return httpresponse.Error(w, httperr(err))
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}
