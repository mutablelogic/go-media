package ffmpeg

import (
	"errors"
	"fmt"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

// decoder for a single stream includes an audio resampler and output
// frame
type decoder struct {
	codec     *ff.AVCodecContext
	resampler *ff.SWRContext
	frame     *ff.AVFrame
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a decoder for a stream
func (r *reader) NewDecoder(media_type MediaType, stream_num int) (*decoder, error) {
	decoder := new(decoder)

	// Find the best stream
	stream_num, codec, err := ff.AVFormat_find_best_stream(r.input, ff.AVMediaType(media_type), stream_num, -1)
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
	stream := r.input.Stream(stream_num)
	if err := ff.AVCodec_parameters_to_context(dec_ctx, stream.CodecPar()); err != nil {
		ff.AVCodec_free_context(dec_ctx)
		return nil, fmt.Errorf("failed to copy codec parameters to decoder context for codec %q", codec.Name())
	}

	// Init the decoder
	if err := ff.AVCodec_open(dec_ctx, dec, nil); err != nil {
		ff.AVCodec_free_context(dec_ctx)
		return nil, err
	} else {
		decoder.codec = dec_ctx
	}

	// Map the decoder
	if _, exists := r.decoders[stream_num]; exists {
		ff.AVCodec_free_context(dec_ctx)
		return nil, fmt.Errorf("decoder for stream %d already exists", stream_num)
	} else {
		r.decoders[stream_num] = decoder
	}

	// Create a frame for decoder output
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		ff.AVCodec_free_context(dec_ctx)
		return nil, errors.New("failed to allocate frame")
	} else {
		decoder.frame = frame
	}

	// Return success
	return decoder, nil
}

// Close the decoder
func (d *decoder) Close() {
	if d.resampler != nil {
		ff.SWResample_free(d.resampler)
	}
	ff.AVUtil_frame_free(d.frame)
	ff.AVCodec_free_context(d.codec)
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d *decoder) String() string {
	return d.codec.String()
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Resample the audio as int16 non-planar samples
// TODO: This should be NewAudioDecoder(..., sample_rate, sample_format, channel_layout)
func (decoder *decoder) ResampleS16Mono(sample_rate int) error {
	// Check decoder type
	if decoder.codec.Codec().Type() != ff.AVMEDIA_TYPE_AUDIO {
		return fmt.Errorf("decoder is not an audio decoder")
	}

	// TODO: Currently hard-coded to 16-bit mono at 44.1kHz
	decoder.frame.SetSampleRate(sample_rate)
	decoder.frame.SetSampleFormat(ff.AV_SAMPLE_FMT_S16)
	if err := decoder.frame.SetChannelLayout(ff.AV_CHANNEL_LAYOUT_STEREO); err != nil {
		return err
	}

	// Create a new resampler
	ctx := ff.SWResample_alloc()
	if ctx == nil {
		return errors.New("failed to allocate resampler")
	} else {
		decoder.resampler = ctx
	}

	// Set options to covert from the codec frame to the decoder frame
	if err := ff.SWResample_set_opts(ctx,
		decoder.frame.ChannelLayout(), decoder.frame.SampleFormat(), decoder.frame.SampleRate(), // destination
		decoder.codec.ChannelLayout(), decoder.codec.SampleFormat(), decoder.codec.SampleRate(), // source
	); err != nil {
		ff.SWResample_free(ctx)
		return err
	}

	// Initialize the resampling context
	if err := ff.SWResample_init(ctx); err != nil {
		ff.SWResample_free(ctx)
		return err
	}

	// Return success
	return nil
}

// Ref:
// https://github.com/romatthe/alephone/blob/b1f7af38b14f74585f0442f1dd757d1238bfcef4/Source_Files/FFmpeg/SDL_ffmpeg.c#L2048

func (decoder *decoder) re(src *ff.AVFrame) (*ff.AVFrame, error) {
	switch decoder.codec.Codec().Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		// Potentially resample the audio
		if decoder.resampler != nil {
			if err := decoder.resample(decoder.frame, src); err != nil {
				return nil, err
			}
			return decoder.frame, nil
		}
	}
	// NO-OP
	return src, nil
}

func (decoder *decoder) resample(dest, src *ff.AVFrame) error {
	//fmt.Println("resample src=>", src)

	dest_samples, err := ff.SWResample_get_out_samples(decoder.resampler, src.NumSamples())
	if err != nil {
		return err
	}
	dest.SetNumSamples(dest_samples)
	dest.SetPts(decoder.get_next_pts(src))

	// Allocate frame buffer
	if err := ff.AVUtil_frame_get_buffer(dest, false); err != nil {
		return err
	} else if err := ff.SWResample_convert_frame(decoder.resampler, src, dest); err != nil {
		return err
	}

	//fmt.Println("in_samples", src.NumSamples(), "out_samples", dest.NumSamples())
	//fmt.Println("in_pts", src.Pts(), "out_pts", dest.Pts())
	//fmt.Println("in_timebase", src.TimeBase(), "out_timebase", dest.TimeBase())

	return nil
}

func (decoder *decoder) get_next_pts(src *ff.AVFrame) int64 {
	ts := src.BestEffortTs()
	if ts == ff.AV_NOPTS_VALUE {
		ts = src.Pts()
	}
	if ts == ff.AV_NOPTS_VALUE {
		return ff.AV_NOPTS_VALUE
	}
	return ff.SWResample_next_pts(decoder.resampler, ts)
}
