package main

import (
	"encoding/json"
	"errors"
	"math"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////

// a wrapper around an output AVStream
type Stream struct {
	// Main parameters
	Codec   *ff.AVCodec
	Encoder *ff.AVCodecContext
	Stream  *ff.AVStream

	tmp_packet       *ff.AVPacket
	next_pts         int64 // pts of the next frame that will be generated
	samples_count    int
	frame            *ff.AVFrame
	tmp_frame        *ff.AVFrame
	packet           *ff.AVPacket
	t, tincr, tincr2 float64
	sws_ctx          *ff.SWSContext
	swr_ctx          *ff.SWRContext
}

func (stream *Stream) String() string {
	data, _ := json.MarshalIndent(stream, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////

// Create a new output stream, add it to the media context and initialize the codec.
func NewStream(ctx *ff.AVFormatContext, codec_id ff.AVCodecID) (*Stream, error) {
	stream := &Stream{}

	// Codec
	codec := ff.AVCodec_find_encoder(codec_id)
	if codec == nil {
		return nil, errors.New("could not find codec")
	} else {
		stream.Codec = codec
	}

	// Packet
	if packet := ff.AVCodec_packet_alloc(); packet == nil {
		return nil, errors.New("could not allocate packet")
	} else {
		stream.tmp_packet = packet
	}

	// Stream
	if str := ff.AVFormat_new_stream(ctx, nil); str == nil {
		ff.AVCodec_packet_free(stream.tmp_packet)
		return nil, errors.New("could not allocate stream")
	} else {
		stream_id := int(ctx.NumStreams())
		stream.Stream = str
		stream.Stream.SetId(stream_id)
	}

	// Codec context
	if encoder := ff.AVCodec_alloc_context(codec); encoder == nil {
		ff.AVCodec_packet_free(stream.tmp_packet)
		return nil, errors.New("could not allocate codec context")
	} else {
		stream.Encoder = encoder
	}

	// Set parameters for the encoder
	switch stream.Codec.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		if fmts := stream.Codec.SampleFormats(); len(fmts) > 0 {
			stream.Encoder.SetSampleFormat(fmts[0])
		} else {
			stream.Encoder.SetSampleFormat(ff.AV_SAMPLE_FMT_FLTP)
		}
		if rates := stream.Codec.SupportedSamplerates(); len(rates) > 0 {
			stream.Encoder.SetSampleRate(rates[0])
		} else {
			stream.Encoder.SetSampleRate(44100)
		}
		stream.Encoder.SetBitRate(64000)
		if err := stream.Encoder.SetChannelLayout(ff.AV_CHANNEL_LAYOUT_STEREO); err != nil {
			ff.AVCodec_packet_free(stream.tmp_packet)
			return nil, err
		}
		stream.Stream.SetTimeBase(ff.AVUtil_rational(1, stream.Encoder.SampleRate()))
	case ff.AVMEDIA_TYPE_VIDEO:
		stream.Encoder.SetBitRate(400000)
		// Resolution must be a multiple of two.
		stream.Encoder.SetWidth(352)
		stream.Encoder.SetHeight(288)
		/* timebase: This is the fundamental unit of time (in seconds) in terms
		 * of which frame timestamps are represented. For fixed-fps content,
		 * timebase should be 1/framerate and timestamp increments should be
		 * identical to 1. */
		stream.Stream.SetTimeBase(ff.AVUtil_rational(1, 25))
		stream.Encoder.SetTimeBase(stream.Stream.TimeBase())
		stream.Encoder.SetGopSize(12) /* emit one intra frame every twelve frames at most */
		stream.Encoder.SetPixFmt(ff.AV_PIX_FMT_YUV420P)

		if stream.Codec.ID() == ff.AV_CODEC_ID_MPEG2VIDEO {
			/* just for testing, we also add B frames */
			stream.Encoder.SetMaxBFrames(2)
		}
		if stream.Codec.ID() == ff.AV_CODEC_ID_MPEG1VIDEO {
			/* Needed to avoid using macroblocks in which some coeffs overflow.
			 * This does not happen with normal video, it just happens here as
			 * the motion of the chroma plane does not match the luma plane. */
			stream.Encoder.SetMbDecision(ff.FF_MB_DECISION_SIMPLE)
		}
	}

	// Some formats want stream headers to be separate
	if ctx.Output().Flags().Is(ff.AVFMT_GLOBALHEADER) {
		stream.Encoder.SetFlags(stream.Encoder.Flags() | ff.AV_CODEC_FLAG_GLOBAL_HEADER)
	}

	// Return success
	return stream, nil
}

func (stream *Stream) Close() {
	ff.AVCodec_packet_free(stream.tmp_packet)
	ff.AVCodec_free_context(stream.Encoder)
	ff.AVUtil_frame_free(stream.frame)
	ff.AVUtil_frame_free(stream.tmp_frame)
	ff.SWResample_free(stream.swr_ctx)
}

func (stream *Stream) Open(opts *ff.AVDictionary) error {
	// Create a copy of the opts
	opt, err := ff.AVUtil_dict_copy(opts, 0)
	if err != nil {
		return err
	}
	defer ff.AVUtil_dict_free(opt)

	// Open the codec
	if err := ff.AVCodec_open(stream.Encoder, stream.Codec, opt); err != nil {
		return err
	}

	switch stream.Codec.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		stream.t = 0
		// increment frequency by 110 Hz per second
		stream.tincr = 2 * math.Pi * 110.0 / float64(stream.Encoder.SampleRate())
		stream.tincr2 = 2 * math.Pi * 110.0 / float64(stream.Encoder.SampleRate()) / float64(stream.Encoder.SampleRate())

		// Number of samples in a frame
		nb_samples := stream.Encoder.FrameSize()
		if stream.Codec.Capabilities().Is(ff.AV_CODEC_CAP_VARIABLE_FRAME_SIZE) {
			nb_samples = 10000
		}

		if frame, err := alloc_audio_frame(stream.Encoder.SampleFormat(), stream.Encoder.ChannelLayout(), stream.Encoder.SampleRate(), nb_samples); err != nil {
			return err
		} else {
			stream.frame = frame
		}
		if frame, err := alloc_audio_frame(ff.AV_SAMPLE_FMT_S16, stream.Encoder.ChannelLayout(), stream.Encoder.SampleRate(), nb_samples); err != nil {
			return err
		} else {
			stream.tmp_frame = frame
		}
		// create resampler context
		if swr_ctx := ff.SWResample_alloc(); swr_ctx == nil {
			return errors.New("could not allocate resample context")
		} else if err := ff.SWResample_set_opts(swr_ctx,
			stream.Encoder.ChannelLayout(), stream.Encoder.SampleFormat(), stream.Encoder.SampleRate(), // out
			stream.Encoder.ChannelLayout(), ff.AV_SAMPLE_FMT_S16, stream.Encoder.SampleRate(), // in
		); err != nil {
			ff.SWResample_free(swr_ctx)
			return err
		} else if err := ff.SWResample_init(swr_ctx); err != nil {
			ff.SWResample_free(swr_ctx)
			return err
		} else {
			stream.swr_ctx = swr_ctx
		}
	case ff.AVMEDIA_TYPE_VIDEO:
		// Allocate a re-usable frame
		if frame, err := alloc_video_frame(stream.Encoder.PixFmt(), stream.Encoder.Width(), stream.Encoder.Height()); err != nil {
			return err
		} else {
			stream.frame = frame
		}
		// If the output format is not YUV420P, then a temporary YUV420P picture is needed too. It is then converted to the required
		// output format.
		if stream.Encoder.PixFmt() != ff.AV_PIX_FMT_YUV420P {
			if frame, err := alloc_video_frame(ff.AV_PIX_FMT_YUV420P, stream.Encoder.Width(), stream.Encoder.Height()); err != nil {
				return err
			} else {
				stream.tmp_frame = frame
			}
		}
	}

	// copy the stream parameters to the muxer
	if err := ff.AVCodec_parameters_from_context(stream.Stream.CodecPar(), stream.Encoder); err != nil {
		return err
	}

	// Return success
	return nil
}

func alloc_video_frame(pix_fmt ff.AVPixelFormat, width, height int) (*ff.AVFrame, error) {
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, errors.New("could not allocate video frame")
	}
	frame.SetWidth(width)
	frame.SetHeight(height)
	frame.SetPixFmt(pix_fmt)

	// allocate the buffers for the frame data
	if err := ff.AVUtil_frame_get_buffer(frame, 0); err != nil {
		ff.AVUtil_frame_free(frame)
		return nil, err
	}

	// Return success
	return frame, nil
}

func alloc_audio_frame(sample_fmt ff.AVSampleFormat, channel_layout ff.AVChannelLayout, sample_rate, nb_samples int) (*ff.AVFrame, error) {
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, errors.New("could not allocate audio frame")
	}
	frame.SetSampleFormat(sample_fmt)
	frame.SetSampleRate(sample_rate)
	frame.SetNumSamples(nb_samples)
	if err := frame.SetChannelLayout(channel_layout); err != nil {
		ff.AVUtil_frame_free(frame)
		return nil, err
	}

	// allocate the buffers for the frame data
	if err := ff.AVUtil_frame_get_buffer(frame, 0); err != nil {
		ff.AVUtil_frame_free(frame)
		return nil, err
	}

	// Return success
	return frame, nil
}
