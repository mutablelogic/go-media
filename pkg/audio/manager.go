package audio

import (
	// Namespace imports
	"errors"
	"fmt"
	"io"

	. "github.com/djthorpe/go-errors"
	"github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type swresample struct{}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *swresample {
	r := new(swresample)
	return r
}

func (r *swresample) Close() error {
	var result error

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *swresample) Convert(in AudioFrame, dest AudioFormat, fn SWResampleFn) error {
	var result error

	// Check arguments
	if in == nil || fn == nil {
		return ErrBadParameter
	}

	// Create a context
	ctx, err := NewContext(in, dest)
	if err != nil {
		return err
	}
	defer ctx.Close()

	// Create an output frame based on input frame
	/*out, err := NewAudioFrame(ctx.DestinationAudioFormat(), in.Duration())
	if err != nil {
		return err
	}*/

	for {
		// Call conversion function once
		if err := fn(ctx.Dest()); err != nil {
			// If error is EOF, then flush the output buffer
			// TODO: Free audio frames
			if err != nil {
				if !errors.Is(err, io.EOF) {
					result = multierror.Append(result, err)
				}
				break
			}
		}
		// Convert the frame
		n, err := ctx.Convert(in)
		fmt.Println("convert n=", n)
		if err != nil {
			fmt.Println("convert err=", err)
		}
	}

	// Return any errors
	return result
}
