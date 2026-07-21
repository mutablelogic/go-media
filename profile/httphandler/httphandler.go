package httphandler

import (
	_ "embed"
	"errors"
	"net/http"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/profile/manager"
	schema "github.com/mutablelogic/go-media/profile/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

//go:embed README.md
var readme []byte

func RegisterCodecHandlers(manager *manager.Profile, router *httprouter.Router) error {
	// Parse the documentation
	documentation := openapi.ParseMarkdown(readme)

	// Add the documentation to the router
	router.Spec().AddTag("Codecs", documentation.Section(2, "Encoders").Body)

	return errors.Join(
		router.Register("codec", nil, func(path httprequest.PathItem) {
			path.Tag("Codecs")

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
			path.Tag("Codecs")

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
				} else if codec := strings.TrimSpace(req.Codec); codec == "" {
					httpresponse.Error(w, gomedia.ErrBadParameter.With("missing required field 'codec'"))
					return
				}

				response, err := manager.CreateAudioProfile(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("Create Audio Profile")
				op.Description(documentation.Section(3, "POST /audio").Body)
				// TODO: Define query parameters and responses for codec profiles
			})
		}),
	)
}
