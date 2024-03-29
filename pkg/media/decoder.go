package media

import (
	// Packages
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	//. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type decoder struct {
	ctx    *ffmpeg.AVCodecContext
	stream *stream
	frame  *frame
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a decoder for a stream
func NewDecoder(stream *stream) *decoder {
	this := new(decoder)

	// Set the stream
	if stream == nil || stream.ctx == nil {
		return nil
	} else {
		this.stream = stream
	}

	// Find the decoder
	decoder := ffmpeg.AVCodec_find_decoder(stream.ctx.CodecPar().CodecID())
	if decoder == nil {
		return nil
	}

	// Allocate a codec context for the decoder
	ctx := ffmpeg.AVCodec_alloc_context3(decoder)
	if ctx == nil {
		return nil
	} else {
		this.ctx = ctx
	}

	// Copy codec parameters from input stream to output codec context
	if err := ffmpeg.AVCodec_parameters_to_context(this.ctx, stream.ctx.CodecPar()); err != nil {
		ffmpeg.AVCodec_free_context_ptr(this.ctx)
		return nil
	}

	// Init the decoders
	if err := ffmpeg.AVCodec_open2(this.ctx, decoder, nil); err != nil {
		ffmpeg.AVCodec_free_context_ptr(this.ctx)
		return nil
	}

	// Create a frame
	if frame := NewFrame(); frame == nil {
		ffmpeg.AVCodec_free_context_ptr(this.ctx)
		return nil
	} else {
		this.frame = frame
	}

	// Return success
	return this
}

func (decoder *decoder) Close() error {
	var result error

	// Release codec context
	if decoder.ctx != nil {
		ffmpeg.AVCodec_free_context_ptr(decoder.ctx)
	}

	// Release frame
	if decoder.frame != nil {
		if err := decoder.frame.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Blank out other fields
	decoder.ctx = nil
	decoder.frame = nil
	decoder.stream = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (decoder *decoder) String() string {
	str := "<media.decoder"
	if decoder.ctx != nil {
		str += fmt.Sprint(" context=", decoder.ctx)
	}
	if decoder.frame != nil {
		str += fmt.Sprint(" frame=", decoder.frame)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (decoder *decoder) AudioFormat() AudioFormat {
	if decoder.ctx == nil {
		return AudioFormat{}
	}
	return AudioFormat{
		Rate:   uint(decoder.ctx.BitRate()),
		Format: fromSampleFormat(decoder.ctx.SampleFormat()),
		Layout: fromChannelLayout(decoder.ctx.ChannelLayout()),
	}
}
