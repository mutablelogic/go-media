package httphandler

import (
	_ "embed"
	"errors"
	"net/http"
	"strings"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/profile/manager"
	schema "github.com/mutablelogic/go-media/profile/schema"
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

func RegisterCapabilityHandlers(manager *manager.Profile, router *httprouter.Router) error {
	// Parse the documentation
	documentation := openapi.ParseMarkdown(readme)

	// Add the documentation to the router
	router.Spec().AddTag("Capabilities", documentation.Section(2, "Container Formats").Body)

	return errors.Join(
		router.Register("containerformat", nil, func(path httprequest.PathItem) {
			path.Tag("Capabilities")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				var req schema.FormatListRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				response, err := manager.ListFormats(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("List Container Formats")
				op.Description(documentation.Section(3, "GET /containerformat").Body)
				op.Query(jsonschema.MustFor[schema.FormatListRequest]())
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.FormatList](), "List of Formats")
			})
		}),
		router.Register("pixelformat", nil, func(path httprequest.PathItem) {
			path.Tag("Capabilities")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				// Request
				var req schema.PixelFormatListRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				// Response
				response, err := manager.ListPixelFormats(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("List Pixel Formats")
				op.Description(documentation.Section(3, "GET /pixelformat").Body)
				op.Query(jsonschema.MustFor[schema.PixelFormatListRequest]())
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.PixelFormatList](), "List of Pixel Formats")
			})
		}),
		router.Register("codec", nil, func(path httprequest.PathItem) {
			path.Tag("Capabilities")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				var req schema.CodecListRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				response, err := manager.ListCodecs(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("List Encoders")
				op.Description(documentation.Section(3, "GET /codec").Body)
				// TODO: Define query parameters and responses for codec profiles
			})
		}),
		router.Register("codec/{name}", nil, func(path httprequest.PathItem) {
			path.Tag("Capabilities")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				response, err := manager.GetCodec(r.Context(), r.PathValue("name"))
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Get Encoder")
				op.Description(documentation.Section(3, "GET /codec/{name}").Body)
				// TODO: Define query parameters and responses for codec profiles
			})
		}),
	)
}

func RegisterAudioProfileHandlers(manager *manager.Profile, router *httprouter.Router) error {
	// Parse the documentation
	documentation := openapi.ParseMarkdown(readme)

	// Add the documentation to the router
	router.Spec().AddTag("Audio Profiles", documentation.Section(2, "Audio Profiles").Body)

	return errors.Join(
		router.Register("audio", nil, func(path httprequest.PathItem) {
			path.Tag("Audio Profiles")

			// POST
			path.Post(func(w http.ResponseWriter, r *http.Request) {
				var req schema.AudioProfileMeta
				if err := httprequest.Read(r, &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				} else if codec := strings.TrimSpace(req.Name); codec == "" {
					httpresponse.Error(w, gomedia.ErrBadParameter.With("missing required field 'codec'"))
					return
				}

				response, err := manager.CreateAudioProfile(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Create Audio Profile")
				op.Description(documentation.Section(3, "POST /audio").Body)
				op.JSONResponse(http.StatusCreated, jsonschema.MustFor[schema.AudioProfile](), "Created Audio Profile")
				op.ErrorResponse(http.StatusNotFound)
			})
		}),
		router.Register("audio/{uuid}", nil, func(path httprequest.PathItem) {
			path.Tag("Audio Profiles")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				if uuid, err := uuid.Parse(r.PathValue("uuid")); err != nil {
					httpresponse.Error(w, gomedia.ErrBadParameter.Withf("invalid uuid: %v", err))
					return
				} else if response, err := manager.GetAudioProfile(r.Context(), uuid); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Get Audio Profile")
				op.Description(documentation.Section(3, "GET /audio/{uuid}").Body)
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.AudioProfile]())
			})

			// DELETE
			path.Delete(func(w http.ResponseWriter, r *http.Request) {
				if uuid, err := uuid.Parse(r.PathValue("uuid")); err != nil {
					httpresponse.Error(w, gomedia.ErrBadParameter.Withf("invalid uuid: %v", err))
					return
				} else if response, err := manager.DeleteAudioProfile(r.Context(), uuid); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Delete Audio Profile")
				op.Description(documentation.Section(3, "DELETE /audio/{uuid}").Body)
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.AudioProfile](), "Deleted Audio Profile")
				op.ErrorResponse(http.StatusNotFound)
			})

			// PATCH
			path.Patch(func(w http.ResponseWriter, r *http.Request) {
				var req schema.AudioProfileMeta
				uuid, err := uuid.Parse(r.PathValue("uuid"))
				if err != nil {
					httpresponse.Error(w, gomedia.ErrBadParameter.Withf("invalid uuid: %v", err))
					return
				} else if err := httprequest.Read(r, &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
					return
				}

				if response, err := manager.UpdateAudioProfile(r.Context(), uuid, req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Update Audio Profile")
				op.Description(documentation.Section(3, "PATCH /audio/{uuid}").Body)
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[schema.AudioProfile](), "Updated Audio Profile")
				op.ErrorResponse(http.StatusNotFound)
			})
		}),
	)
}
