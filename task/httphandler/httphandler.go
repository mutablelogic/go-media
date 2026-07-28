package httphandler

import (
	_ "embed"
	"errors"
	"net/http"
	"slices"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/task/manager"
	schema "github.com/mutablelogic/go-media/task/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

//go:embed README.md
var readme []byte

func RegisterTaskHandlers(manager *manager.Manager, router *httprouter.Router) error {
	// Parse the documentation
	documentation := openapi.ParseMarkdown(readme)

	// Add the documentation to the router
	router.Spec().AddTag("Tasks", documentation.Section(1, "Tasks").Body)

	return errors.Join(
		router.Register("task", nil, func(path httprequest.PathItem) {
			path.Tag("Tasks")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				var req schema.TaskListRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				response, err := manager.ListTasks(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("List Tasks")
				op.Description(documentation.Section(2, "GET /task").Body)
				op.Query(jsonschema.MustFor[schema.TaskListRequest]())
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.TaskList](), "List of Tasks")
			})
		}),
		router.Register("task/event", nil, func(path httprequest.PathItem) {
			path.Tag("Tasks")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				var req schema.TaskEventRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				var id uuid.UUID
				if req.UUID != nil {
					parsed, err := uuid.Parse(*req.UUID)
					if err != nil {
						httpresponse.Error(w, gomedia.HTTPErr(gomedia.ErrBadParameter.Withf("invalid uuid: %v", err)))
						return
					}
					id = parsed
				}

				stream := httpresponse.NewTextStream(w)
				if stream == nil {
					httpresponse.Error(w, gomedia.HTTPErr(gomedia.ErrInternalError.With("failed to open event stream")))
					return
				}
				defer stream.Close()

				// Blocks until the client disconnects (r.Context() is done)
				// or the task manager itself stops.
				_ = manager.Subscribe(r.Context(), func(e *schema.Event) {
					if req.UUID != nil && e.Status.UUID != id {
						return
					}
					if len(req.Event) > 0 && !slices.Contains(req.Event, e.Name) {
						return
					}
					stream.Write(string(e.Name), e.Status)
				})
			}, func(op httprequest.PathOperation) {
				op.Summary("Subscribe to Task Events")
				op.Description(documentation.Section(2, "GET /task/event").Body)
				op.Query(jsonschema.MustFor[schema.TaskEventRequest]())
				op.Response(http.StatusOK, "text/event-stream", "Stream of task events")
			})
		}),
		router.Register("task/{uuid}", nil, func(path httprequest.PathItem) {
			path.Tag("Tasks")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				if id, err := uuid.Parse(r.PathValue("uuid")); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(gomedia.ErrBadParameter.Withf("invalid uuid: %v", err)))
					return
				} else if response, err := manager.GetTask(r.Context(), id); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Get Task")
				op.Description(documentation.Section(2, "GET /task/{uuid}").Body)
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.Status]())
				op.ErrorResponse(http.StatusNotFound)
			})

			// DELETE
			path.Delete(func(w http.ResponseWriter, r *http.Request) {
				id, err := uuid.Parse(r.PathValue("uuid"))
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(gomedia.ErrBadParameter.Withf("invalid uuid: %v", err)))
					return
				}
				if err := manager.Cancel(r.Context(), id); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}
				if response, err := manager.GetTask(r.Context(), id); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Cancel Task")
				op.Description(documentation.Section(2, "DELETE /task/{uuid}").Body)
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.Status](), "Cancelled Task")
				op.ErrorResponse(http.StatusNotFound)
			})
		}),
	)
}
