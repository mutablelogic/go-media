package httphandler

import (
	"net/http"

	// Packages
	task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterHandlers(router *http.ServeMux, prefix string, manager *task.Manager) {
	// TODO
}
