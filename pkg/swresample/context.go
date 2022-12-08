package swresample

import (
	"runtime"

	// Package imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

// Package imports

////////////////////////////////////////////////////////////////////////////////
// TYPES

type swcontext struct {
	ctx *ffmpeg.SWRContext
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new empty context object
func NewContext() *swcontext {
	r := new(swcontext)
	r.ctx = ffmpeg.SWR_alloc()
	runtime.SetFinalizer(r, finalizer)
	return r
}

// Free resources associated with the context
func (r *swcontext) Close() error {
	var result error

	// Free context
	if r.ctx != nil {
		r.ctx.SWR_free()
		r.ctx = nil
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *swcontext) initialize() error {
	// Check parameters
	if r.ctx == nil {
		return ErrInternalAppError.With("Context is closed")
	} else if r.ctx.SWR_is_initialized() {
		return ErrOutOfOrder.With("Context already initialized")
	}
	return r.ctx.SWR_init()
}

func finalizer(r *swcontext) {
	if r.ctx != nil {
		panic("swresample: context not closed")
	}
}
