package audio

import (
	// Namespace imports
	. "github.com/djthorpe/go-errors"
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

func (r *swresample) Convert(in AudioFrame, out AudioFormat, fn SWResampleFn) error {
	var result error

	// Check arguments
	if in == nil || fn == nil {
		return ErrBadParameter
	}

	// Create a context
	ctx, err := NewContext(in, out)
	if err != nil {
		return err
	}
	defer ctx.Close()

	// Create an output frame based on input frame
	out,err := NewAudioFrame(ctx.DestinationAudioFormat(), in.Duration())
	if err != nil {
		return err
	}

	// Call conversion function once
	if err := fn(ctx); err != nil {

	/*
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
	*/

	// Return any errors
	return result
}

/*

 * Once all values have been set, it must be initialized with swr_init(). If
 * you need to change the conversion parameters, you can change the parameters
 * using @ref AVOptions, as described above in the first example; or by using
 * swr_alloc_set_opts2(), but with the first argument the allocated context.
 * You must then call swr_init() again.
 *
 * The conversion itself is done by repeatedly calling swr_convert().
 * Note that the samples may get buffered in swr if you provide insufficient
 * output space or if sample rate conversion is done, which requires "future"
 * samples. Samples that do not require future input can be retrieved at any
 * time by using swr_convert() (in_count can be set to 0).
 * At the end of conversion the resampling buffer can be flushed by calling
 * swr_convert() with NULL in and 0 in_count.
 *
 * The samples used in the conversion process can be managed with the libavutil
 * @ref lavu_sampmanip "samples manipulation" API, including av_samples_alloc()
 * function used in the following example.
 *
 * The delay between input and output, can at any time be found by using
 * swr_get_delay().
 *
 * The following code demonstrates the conversion loop assuming the parameters
 * from above and caller-defined functions get_input() and handle_output():
 * @code
 * uint8_t **input;
 * int in_samples;
 *
 * while (get_input(&input, &in_samples)) {
 *     uint8_t *output;
 *     int out_samples = av_rescale_rnd(swr_get_delay(swr, 48000) +
 *                                      in_samples, 44100, 48000, AV_ROUND_UP);
 *     av_samples_alloc(&output, NULL, 2, out_samples,
 *                      AV_SAMPLE_FMT_S16, 0);
 *     out_samples = swr_convert(swr, &output, out_samples,
 *                                      input, in_samples);
 *     handle_output(output, out_samples);
 *     av_freep(&output);
 * }
 * @endcode
 *
 * When the conversion is finished, the conversion
 * context and everything associated with it must be freed with swr_free().
 * A swr_close() function is also available, but it exists mainly for
 * compatibility with libavresample, and is not required to be called.
 *
 * There will be no memory leak if the data is not completely flushed before
 * swr_free().
 */
