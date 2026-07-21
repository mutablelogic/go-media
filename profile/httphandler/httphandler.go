package httphandler

import (
	_ "embed"
	"errors"
	"net/http"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/profile/manager"
	"github.com/mutablelogic/go-media/profile/schema"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

//go:embed README.md
var readme []byte

func RegisterHandlers(manager *manager.Profile, router *httprouter.Router) error {
	// Parse the documentation
	documentation := openapi.ParseMarkdown(readme)

	// Add the documentation to the router
	router.Spec().AddTag("Codecs", documentation.Section(1, "Codecs").Body)

	return errors.Join(
		router.Register("codec", nil, func(path httprequest.PathItem) {
			path.Tag("Codecs")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				var req schema.CodecListRequest
				if err := httprequest.Query(r.URL.Query(), &req); err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				}

				response, err := manager.ListCodecs(r.Context(), req)
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("List Encoders")
				op.Description(documentation.Section(2, "GET /codec").Body)
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
				op.Description(documentation.Section(2, "GET /codec/{name}").Body)
				// TODO: Define query parameters and responses for codec profiles
			})
		}),
	)
}
