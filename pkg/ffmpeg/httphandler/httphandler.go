package httphandler

import (
	"errors"
	"net/http"

	// Packages
	task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterHandlers(router *http.ServeMux, prefix string, manager *task.Manager) {
	RegisterAudioChannelLayoutHandlers(router, prefix, manager)
	RegisterCodecHandlers(router, prefix, manager)
	RegisterFilterHandlers(router, prefix, manager)
	RegisterFormatHandlers(router, prefix, manager)
	RegisterPixelFormatHandlers(router, prefix, manager)
	RegisterSampleFormatHandlers(router, prefix, manager)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// httperr converts pg errors to appropriate HTTP errors.
// Returns the original error if it's already an httpresponse.Err,
func httperr(err error) error {
	if err == nil {
		return nil
	}

	// If already an HTTP error, return as-is
	var httpErr httpresponse.Err
	if errors.As(err, &httpErr) {
		return err
	}

	// TODO: Map ffmpeg errors to HTTP errors
	return httpresponse.ErrInternalError.With(err.Error())
}
