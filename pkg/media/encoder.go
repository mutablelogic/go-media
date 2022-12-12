package media

import (
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	//. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type encoder struct {
	codec *ffmpeg.AVCodec
	ctx   *ffmpeg.AVCodecContext
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a encoder for a stream
func NewEncoderByName(name string) *encoder {
	this := new(encoder)

	// Find encoder
	if codec := ffmpeg.AVCodec_find_encoder_by_name(name); codec == nil {
		return nil
	} else {
		this.codec = codec
	}

	// Allocate context
	if ctx := ffmpeg.AVCodec_alloc_context3(this.codec); ctx == nil {
		return nil
	} else {
		this.ctx = ctx
	}

	// Return success
	return this
}

func (encoder *encoder) Close() error {
	var result error

	// Release context
	if encoder.ctx != nil {
		ffmpeg.AVCodec_free_context_ptr(encoder.ctx)
	}

	// Blank out other fields
	encoder.ctx = nil
	encoder.codec = nil

	// Return success
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (encoder *encoder) String() string {
	str := "<media.encoder"
	if encoder.codec != nil {
		str += " name=" + encoder.codec.Name()
	}
	return str + ">"
}

/*

            // In this example, we transcode to same properties (picture size,
             // sample rate etc.). These properties can be changed for output
             // streams easily using filters
			 if (dec_ctx->codec_type == AVMEDIA_TYPE_VIDEO) {
                enc_ctx->height = dec_ctx->height;
                enc_ctx->width = dec_ctx->width;
                enc_ctx->sample_aspect_ratio = dec_ctx->sample_aspect_ratio;
                // take first format from list of supported formats
                if (encoder->pix_fmts)
                    enc_ctx->pix_fmt = encoder->pix_fmts[0];
                else
                    enc_ctx->pix_fmt = dec_ctx->pix_fmt;
                // video time_base can be set to whatever is handy and supported by encoder
            } else {
                enc_ctx->sample_rate = dec_ctx->sample_rate;
                ret = av_channel_layout_copy(&enc_ctx->ch_layout, &dec_ctx->ch_layout);
                if (ret < 0)
                    return ret;
                // take first format from list of supported formats
                enc_ctx->sample_fmt = encoder->sample_fmts[0];
                enc_ctx->time_base = (AVRational){1, enc_ctx->sample_rate};
            }

            if (ofmt_ctx->oformat->flags & AVFMT_GLOBALHEADER)
                enc_ctx->flags |= AV_CODEC_FLAG_GLOBAL_HEADER;

            // Third parameter can be used to pass settings to encoder
            ret = avcodec_open2(enc_ctx, encoder, NULL);
            if (ret < 0) {
                av_log(NULL, AV_LOG_ERROR, "Cannot open video encoder for stream #%u\n", i);
                return ret;
            }
            ret = avcodec_parameters_from_context(out_stream->codecpar, enc_ctx);
            if (ret < 0) {
                av_log(NULL, AV_LOG_ERROR, "Failed to copy encoder parameters to output stream #%u\n", i);
                return ret;
            }

            out_stream->time_base = enc_ctx->time_base;
            stream_ctx[i].enc_ctx = enc_ctx;
*/
