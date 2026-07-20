package httphandler

import (
	_ "embed"
	"errors"
	"net/http"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	manager "github.com/mutablelogic/go-media/profile/manager"
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
	router.Spec().AddTag("Codec Profiles", documentation.Section(1, "Codec Profiles").Body)

	return errors.Join(
		router.Register("codec", nil, func(path httprequest.PathItem) {
			path.Tag("Codec Profiles")

			// GET
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				codecs, err := manager.ListAudioCodecs(r.Context())
				if err != nil {
					httpresponse.Error(w, gomedia.HTTPErr(err))
				} else {
					httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), codecs)
				}
			}, func(op httprequest.PathOperation) {
				op.Summary("List Codecs")
				op.Description(documentation.Section(2, "GET /codec").Body)
				// TODO: Define query parameters and responses for codec profiles
			})
		}),
	)
}
