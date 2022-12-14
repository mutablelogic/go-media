package audio

import (
	"errors"
	"fmt"
	"io"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

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
	var dest AudioFrame

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

	// Calculate the number of samples in the output buffer
	n_samples := ctx.GetSrcDelay() + int64(in.Samples())
	out_samples := ffmpeg.AVUtil_av_rescale_rnd(n_samples, int64(ctx.DestAudioFormat().Rate), int64(ctx.SrcAudioFormat().Rate), ffmpeg.AV_ROUND_UP)
	fmt.Println("out_samples=", out_samples)

	for {
		// Callback function to get input
		if err := fn(dest); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			result = multierror.Append(result, err)
			break
		}
		if !in.IsPlanar() && !out.IsPlanar() {

		}
		ch := in.Channels()
		if err := ffmpeg.SWR_convert(); err != nil {
			result = multierror.Append(result, err)
			break
		}
	}

	/*
		// Repeat calling conversion until error
		   	var in, out []byte
		   	var err error
		   	var n int

		   FOR_LOOP:

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
