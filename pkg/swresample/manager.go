package swresample

import (
	// Package imports
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type swresample struct {
	ctx []*swcontext
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *swresample {
	r := new(swresample)
	return r
}

func (r *swresample) Close() error {
	var result error

	// Free all contexts
	for _, ctx := range r.ctx {
		if ctx != nil {
			if err := ctx.Close(); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *swresample) NewContext() *swcontext {
	ctx := NewContext()
	r.ctx = append(r.ctx, ctx)
	return ctx
}

func (r *swresample) ConvertBytes(ctx SWResampleContext, fn SWResampleConvertBytes) error {
	if err := ctx.(*swcontext).initialize(); err != nil {
		return err
	}
	return nil
}
