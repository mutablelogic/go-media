package ffmpeg

import (
	"fmt"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

// decoder for a single stream
type decoder struct {
	*ff.AVCodecContext
}

// Create a decoder for a stream
func (r *reader) NewDecoder(media_type MediaType, stream_num int) (*decoder, error) {
	decoder := new(decoder)

	// Find the best stream
	stream_num, codec, err := ff.AVFormat_find_best_stream(r.AVFormatContext, ff.AVMediaType(media_type), stream_num, -1)
	if err != nil {
		return nil, err
	}

	// Find the decoder for the stream
	dec := ff.AVCodec_find_decoder(codec.ID())
	if dec == nil {
		return nil, fmt.Errorf("failed to find decoder for codec %q", codec.Name())
	}

	// Allocate a codec context for the decoder
	dec_ctx := ff.AVCodec_alloc_context(dec)
	if dec_ctx == nil {
		return nil, fmt.Errorf("failed to allocate codec context for codec %q", codec.Name())
	}

	// Copy codec parameters from input stream to output codec context
	stream := r.AVFormatContext.Stream(stream_num)
	if err := ff.AVCodec_parameters_to_context(dec_ctx, stream.CodecPar()); err != nil {
		ff.AVCodec_free_context(dec_ctx)
		return nil, fmt.Errorf("failed to copy codec parameters to decoder context for codec %q", codec.Name())
	}

	// Init the decoder
	if err := ff.AVCodec_open(dec_ctx, dec, nil); err != nil {
		ff.AVCodec_free_context(dec_ctx)
		return nil, err
	} else {
		decoder.AVCodecContext = dec_ctx
	}

	// Map the decoder
	if _, exists := r.decoders[stream_num]; exists {
		ff.AVCodec_free_context(dec_ctx)
		return nil, fmt.Errorf("decoder for stream %d already exists", stream_num)
	} else {
		r.decoders[stream_num] = decoder
	}

	// Return success
	return decoder, nil
}

// Close the decoder
func (d *decoder) Close() {
	ff.AVCodec_free_context(d.AVCodecContext)
}
