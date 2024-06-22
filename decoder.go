package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

// demuxer context - deconstructs media into packets
type demuxer struct {
	input    *ff.AVFormatContext
	decoders map[int]*decoder
	frame    *ff.AVFrame // Source frame
}

// decoder context - decodes packets into frames
type decoder struct {
	stream    int
	codec     *ff.AVCodecContext
	frame     *ff.AVFrame    // Destination frame
	reframe   *ff.AVFrame    // Destination frame after resample or resize
	resampler *ff.SWRContext // Resampler for audio
	rescaler  *ff.SWSContext // Rescaler for video
}

var _ Decoder = (*demuxer)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newDemuxer(input *ff.AVFormatContext, mapfn DecoderMapFunc) (*demuxer, error) {
	demuxer := new(demuxer)
	demuxer.input = input
	demuxer.decoders = make(map[int]*decoder)

	// Get all the streams
	streams := input.Streams()

	// Use standard map function if none provided
	if mapfn == nil {
		mapfn = func(stream Stream) (Parameters, error) {
			return stream.Parameters(), nil
		}
	}

	// Create a decoder for each stream
	// The decoder map function should be returning the parameters for the
	// destination frame. If it's nil then it's mostly a copy.
	var result error
	for _, stream := range streams {
		// Get decoder parameters and map to a decoder
		parameters, err := mapfn(newStream(stream))
		if err != nil {
			result = errors.Join(result, err)
		} else if parameters == nil {
			continue
		} else if decoder, err := demuxer.newDecoder(stream, parameters); err != nil {
			result = errors.Join(result, err)
		} else {
			streamNum := stream.Index()
			demuxer.decoders[streamNum] = decoder
		}
	}

	// Return any errors
	if result != nil {
		return nil, errors.Join(result, demuxer.close())
	}

	// Create a frame for encoding - after resampling and resizing
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		return nil, errors.Join(demuxer.close(), errors.New("failed to allocate frame"))
	} else {
		demuxer.frame = frame
	}

	// Return success
	return demuxer, nil
}

func (d *demuxer) newDecoder(stream *ff.AVStream, dest Parameters) (*decoder, error) {
	decoder := new(decoder)
	decoder.stream = stream.Id()

	// Use parameters to create the decoder resampler or resizer
	src := stream.CodecPar()
	if equals, err := equalsStream(dest, src); err != nil {
		return nil, err
	} else if !equals {
		switch src.CodecType() {
		case ff.AVMEDIA_TYPE_AUDIO:
			fmt.Println("TODO: set up resampler", dest)
		case ff.AVMEDIA_TYPE_VIDEO:
			if rescaler, frame, err := newResizer(dest, src); err != nil {
				return nil, err
			} else {
				decoder.rescaler = rescaler
				decoder.reframe = frame
			}
		default:
			return nil, fmt.Errorf("new decoder: unsupported stream type %v", src.CodecType())
		}
	}

	// Create a codec context for the decoder
	codec := ff.AVCodec_find_decoder(stream.CodecPar().CodecID())
	if codec == nil {
		return nil, fmt.Errorf("failed to find decoder for codec %q", stream.CodecPar().CodecID())
	} else if ctx := ff.AVCodec_alloc_context(codec); ctx == nil {
		return nil, fmt.Errorf("failed to allocate codec context for codec %q", codec.Name())
	} else {
		decoder.codec = ctx
	}

	// Copy codec parameters from input stream to output codec context
	if err := ff.AVCodec_parameters_to_context(decoder.codec, stream.CodecPar()); err != nil {
		return nil, errors.Join(decoder.close(), fmt.Errorf("failed to copy codec parameters to decoder context for codec %q", codec.Name()))
	}

	// Init the decoder
	if err := ff.AVCodec_open(decoder.codec, codec, nil); err != nil {
		return nil, errors.Join(decoder.close(), err)
	}

	// Create a frame for decoder output - after resize/resample
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		return nil, errors.Join(decoder.close(), errors.New("failed to allocate frame"))
	} else {
		decoder.frame = frame
	}

	// Return success
	return decoder, nil
}

func newResizer(dest Parameters, src *ff.AVCodecParameters) (*ff.SWSContext, *ff.AVFrame, error) {
	// Get native pixel format
	dest_pixel_format := ff.AVUtil_get_pix_fmt(dest.PixelFormat())

	// Create scaling context and destination frame
	if ctx := ff.SWScale_get_context(
		src.Width(), src.Height(), src.PixelFormat(), // source
		dest.Width(), dest.Height(), dest_pixel_format, // destination
		ff.SWS_BILINEAR, nil, nil, nil); ctx == nil {
		return nil, nil, errors.New("failed to allocate swscale context")
	} else if frame := ff.AVUtil_frame_alloc(); frame == nil {
		ff.SWScale_free_context(ctx)
		return nil, nil, errors.New("failed to allocate frame")
	} else {
		return ctx, frame, nil
	}
}

func (d *demuxer) close() error {
	var result error

	// Free decoded frame
	if d.frame != nil {
		ff.AVUtil_frame_free(d.frame)
	}

	// Free resources
	for _, decoder := range d.decoders {
		result = errors.Join(result, decoder.close())
	}
	d.decoders = nil

	// Return any errors
	return result
}

func (d *decoder) close() error {
	var result error

	// Free the codec context
	if d.codec != nil {
		ff.AVCodec_free_context(d.codec)
	}

	// Free the resampler
	if d.resampler != nil {
		ff.SWResample_free(d.resampler)
	}

	// Free the rescaler
	if d.rescaler != nil {
		ff.SWScale_free_context(d.rescaler)
	}

	// Free rescaler frame
	if d.reframe != nil {
		ff.AVUtil_frame_free(d.reframe)
	}

	// Free destination frame
	if d.frame != nil {
		ff.AVUtil_frame_free(d.frame)
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *demuxer) Demux(ctx context.Context, fn DecoderFunc) error {
	if fn == nil {
		return errors.New("no decoder function provided")
	}
	return d.demux(ctx, fn, nil)
}

func (d *demuxer) Decode(ctx context.Context, fn FrameFunc) error {
	if fn == nil {
		return errors.New("no decoder function provided")
	}
	return d.demux(ctx, nil, fn)
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (d *demuxer) demux(ctx context.Context, demuxfn DecoderFunc, framefn FrameFunc) error {
	// Allocate a packet
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		return errors.New("failed to allocate packet")
	}
	defer ff.AVCodec_packet_free(packet)

	// Read packets
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		default:
			if err := ff.AVFormat_read_frame(d.input, packet); errors.Is(err, io.EOF) {
				break FOR_LOOP
			} else if err != nil {
				return err
			}
			stream := packet.StreamIndex()
			if decoder := d.decoders[stream]; decoder != nil {
				if err := decoder.decode(packet, demuxfn, framefn); errors.Is(err, io.EOF) {
					break FOR_LOOP
				} else if err != nil {
					return err
				}
			}
			// Unreference the packet
			ff.AVCodec_packet_unref(packet)
		}
	}

	// Flush the decoders
	for _, decoder := range d.decoders {
		if err := decoder.decode(nil, demuxfn, framefn); err != nil {
			return err
		}
	}

	// Return the context error - will be cancelled, perhaps, or nil if the
	// demuxer finished successfully without cancellation
	return ctx.Err()
}

func (d *decoder) decode(packet *ff.AVPacket, demuxfn DecoderFunc, framefn FrameFunc) error {
	if demuxfn != nil {
		// Send the packet to the user defined packet function
		return demuxfn(newPacket(packet))
	}

	// Submit the packet to the decoder (nil packet will flush the decoder)
	if err := ff.AVCodec_send_packet(d.codec, packet); err != nil {
		return err
	}

	// get all the available frames from the decoder
	var result error
	for {
		if err := ff.AVCodec_receive_frame(d.codec, d.frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished decoding packet or EOF
			break
		} else if err != nil {
			return err
		}

		// Resample or resize the frame, then pass to the frame function
		if frame, err := d.re(d.frame); err != nil {
			return err
		} else if err := framefn(newFrame(frame)); errors.Is(err, io.EOF) {
			// End early, return EOF
			result = io.EOF
			break
		} else if err != nil {
			return err
		}
	}

	// Flush the resizer or resampler if we haven't received an EOF
	if result == nil {
		if frame, err := d.re(nil); err != nil {
			return err
		} else if frame == nil {
			// NOOP
		} else if err := framefn(newFrame(d.frame)); errors.Is(err, io.EOF) {
			// NOOP
		} else if err != nil {
			return err
		}
	}

	// Return success or EOF
	return result
}

func (d *decoder) re(src *ff.AVFrame) (*ff.AVFrame, error) {
	switch d.codec.Codec().Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		fmt.Println("TODO: resample audio", src)
	case ff.AVMEDIA_TYPE_VIDEO:
		if d.rescaler != nil && src != nil {
			// Rescale the video
			if err := rescale(d.rescaler, d.reframe, src); err != nil {
				return nil, err
			} else {
				return d.reframe, nil
			}
		}
	}

	//	if err := decoder.rescale(decoder.frame, src); err != nil {
	//		return nil, err
	//	}

	// NO-OP - just return the source frame
	return src, nil
}

func rescale(ctx *ff.SWSContext, dest, src *ff.AVFrame) error {
	// Copy properties from source
	if err := ff.AVUtil_frame_copy_props(dest, src); err != nil {
		return fmt.Errorf("failed to copy props: %w", err)
	}
	// Perform rescale
	if err := ff.SWScale_scale_frame(ctx, dest, src, false); err != nil {
		return fmt.Errorf("SWScale_scale_frame: %w", err)
	}
	return nil
}

// Return an error if the parameters don't match the stream type (AUDIO, VIDEO)
// Return true if the codec parameters are compatible with the stream
func equalsStream(dest Parameters, src *ff.AVCodecParameters) (bool, error) {
	switch src.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		if !dest.Type().Is(AUDIO) {
			return false, fmt.Errorf("source is audio, but destination is %v", dest.Type())
		} else {
			return equalsAudioPar(dest, src), nil
		}
	case ff.AVMEDIA_TYPE_VIDEO:
		if !dest.Type().Is(VIDEO) {
			return false, fmt.Errorf("source is video, but destination are %v", dest.Type())
		} else {
			return equalsVideoPar(dest, src), nil
		}
	default:
		return false, fmt.Errorf("unsupported source %v", src.CodecType())
	}
}

// Return true if the audio parameters are compatible with the stream
func equalsAudioPar(parameters Parameters, codec *ff.AVCodecParameters) bool {
	samplefmt := ff.AVUtil_get_sample_fmt_name(codec.SampleFormat())
	if samplefmt != parameters.SampleFormat() {
		return false
	}
	ch_layout := ff.AVChannelLayout(codec.ChannelLayout())
	channellayout, err := ff.AVUtil_channel_layout_describe(&ch_layout)
	if err != nil || channellayout != parameters.ChannelLayout() {
		return false
	}
	if codec.Samplerate() != parameters.Samplerate() {
		return false
	}
	// Matches
	return true
}

// Return true if the video parameters are compatible with the stream
func equalsVideoPar(parameters Parameters, codec *ff.AVCodecParameters) bool {
	pixelfmt := ff.AVUtil_get_pix_fmt_name(codec.PixelFormat())
	if pixelfmt != parameters.PixelFormat() {
		return false
	}
	if codec.Width() != parameters.Width() {
		return false
	}
	if codec.Height() != parameters.Height() {
		return false
	}
	// Matches
	return true
}
