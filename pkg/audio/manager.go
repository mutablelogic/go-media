package audio

import (
	"fmt"
	"io"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
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
	// Initialize the context
	if err := ctx.(*swcontext).initialize(); err != nil {
		return err
	}
	// Repeat calling conversion until error
	var in, out []byte
	var err error
	var n int
FOR_LOOP:
	for {
		// Call to get an input buffer
		in, err = fn(ctx, out)
		if err != nil {
			break FOR_LOOP
		}
		n, err := ctx.(*swcontext).ctx.SWR_convert_bytes(out, in)
		if err != nil {
			break FOR_LOOP
		}
		fmt.Println("n=", n)
	}
	// If error is EOF, then flush the output buffer
	if err == io.EOF {
		n, err = ctx.(*swcontext).ctx.SWR_convert_bytes(out, nil)
		if err == nil {
			_, err = fn(ctx, out)
		}
		fmt.Println("EOF n=", n)
	}
	// If error is EOF, then return nil
	if err == io.EOF {
		return nil
	} else {
		return err
	}
}
