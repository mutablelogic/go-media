package ffmpeg

import (
	"context"
	"errors"
	"io"
	"sync"
	"syscall"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Return parameters if a stream should be decoded and either resampled or
// resized. Return nil if you want to ignore the stream, or pass back the
// stream parameters if you want to copy the stream without any changes.
type DecoderMapFunc func(int, *Par) (*Par, error)

// DecoderFrameFn is a function which is called to send a frame after decoding. It should
// return nil to continue decoding or io.EOF to stop.
type DecoderFrameFn func(int, *Frame) error

// DecoderSubtitleFn is a function which is called to send a decoded subtitle. It should
// return nil to continue decoding or io.EOF to stop.
type DecoderSubtitleFn func(int, *ff.AVSubtitle) error

// DecoderPacketFn is a function which is called to send a packet after demuxing. It should
// return nil to continue reading packets or io.EOF to stop. Use this for stream copying
// without decode/encode overhead.
type DecoderPacketFn func(int, *Packet) error

// Decoder manages the state for decoding packets and frames from a reader
type decoder struct {
	mu       sync.Mutex
	reader   *Reader
	pkt      *ff.AVPacket
	busy     bool
	decoders map[int]*streamDecoder
}

// Per-stream decoder that handles individual stream decoding
type streamDecoder struct {
	stream    int
	codec     *ff.AVCodecContext
	resampler *Resampler
	frame     *ff.AVFrame
	timeBase  ff.AVRational
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new decoder for the reader
func newDecoder(r *Reader) (*decoder, error) {
	d := &decoder{
		reader: r,
	}

	// Allocate packet for reading - this will be reused for each read
	d.pkt = ff.AVCodec_packet_alloc()
	if d.pkt == nil {
		return nil, errors.New("failed to allocate packet")
	}

	return d, nil
}

// Free decoder resources
func (d *decoder) free() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Free all stream decoders
	for _, dec := range d.decoders {
		dec.close()
	}
	d.decoders = nil

	if d.pkt != nil {
		ff.AVCodec_packet_free(d.pkt)
		d.pkt = nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Read packets without decoding to frames. The packetfn is called for each packet
// read from any stream. Use this for stream copying or remuxing without transcoding.
func (d *decoder) readPackets(ctx context.Context, packetfn DecoderPacketFn) error {
	d.mu.Lock()
	if d.busy {
		d.mu.Unlock()
		return errors.New("decoder already in use")
	}
	if d.pkt == nil {
		d.mu.Unlock()
		return errors.New("decoder is closed")
	}
	d.busy = true
	d.mu.Unlock()

	// Ensure we clear busy flag on return
	defer func() {
		d.mu.Lock()
		d.busy = false
		d.mu.Unlock()
	}()

	// Read packets until EOF or error
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Unref any previous packet data before reading
		ff.AVCodec_packet_unref(d.pkt)

		// Read next packet from any stream
		if err := ff.AVFormat_read_frame(d.reader.input, d.pkt); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		// Wrap the packet and call user function
		packet := newPacket(d.pkt)
		if err := packetfn(packet.Stream(), packet); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

// Decode and demux the media stream into frames and subtitles. The map function determines which
// streams to decode and what output parameters to use. The framefn is called for each
// decoded frame from any mapped stream. The subtitlefn is called for each decoded subtitle.
func (d *decoder) decodeFrames(ctx context.Context, mapfn DecoderMapFunc, framefn DecoderFrameFn, subtitlefn DecoderSubtitleFn) error {
	d.mu.Lock()
	if d.busy {
		d.mu.Unlock()
		return errors.New("decoder already in use")
	}
	if d.pkt == nil {
		d.mu.Unlock()
		return errors.New("decoder is closed")
	}
	d.busy = true
	d.mu.Unlock()

	// Ensure we clear busy flag on return
	defer func() {
		d.mu.Lock()
		d.busy = false
		d.mu.Unlock()
	}()

	// Map streams to decoders
	if err := d.mapStreams(mapfn); err != nil {
		return err
	}

	// Check that we have at least one decoder
	if len(d.decoders) == 0 {
		return errors.New("no streams to decode")
	}

	// Read and decode packets
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Unref any previous packet data before reading
		ff.AVCodec_packet_unref(d.pkt)

		// Read next packet from any stream
		if err := ff.AVFormat_read_frame(d.reader.input, d.pkt); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		// Get the stream decoder for this packet
		streamIndex := d.pkt.StreamIndex()
		dec := d.decoders[streamIndex]
		if dec == nil {
			// Skip packets from unmapped streams
			continue
		}

		// Decode this packet
		if err := dec.decode(d.pkt, framefn, subtitlefn); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}

	// Flush all decoders
	for streamIndex, dec := range d.decoders {
		if err := dec.decode(nil, framefn, subtitlefn); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		// Mark this decoder as flushed by removing it
		delete(d.decoders, streamIndex)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Map streams to decoders based on the map function
func (d *decoder) mapStreams(fn DecoderMapFunc) error {
	// Initialize decoders map
	d.decoders = make(map[int]*streamDecoder)

	// Default map function copies all streams
	if fn == nil {
		fn = func(_ int, par *Par) (*Par, error) {
			return par, nil
		}
	}

	// Create a decoder for each stream
	var result error
	for _, stream := range d.reader.input.Streams() {
		streamIndex := stream.Index()

		// Get decoder parameters and map to a decoder
		srcPar := &Par{
			AVCodecParameters: *stream.CodecPar(),
			timebase:          stream.TimeBase(),
		}

		destPar, err := fn(streamIndex, srcPar)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		if destPar == nil {
			// Stream is not mapped (ignored)
			continue
		}

		// Create decoder for this stream
		dec, err := newStreamDecoder(stream, srcPar, destPar, d.reader.force)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}

		// Check for duplicate
		if _, exists := d.decoders[streamIndex]; exists {
			result = errors.Join(result, errors.New("duplicate stream index"))
			dec.close()
			continue
		}

		d.decoders[streamIndex] = dec
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - STREAM DECODER

// Create a new stream decoder
func newStreamDecoder(stream *ff.AVStream, srcPar, destPar *Par, force bool) (*streamDecoder, error) {
	dec := &streamDecoder{
		stream:   stream.Index(),
		timeBase: stream.TimeBase(),
	}

	// Allocate frame for decoder output
	dec.frame = ff.AVUtil_frame_alloc()
	if dec.frame == nil {
		return nil, errors.New("failed to allocate frame")
	}

	// Find and allocate codec
	codec := ff.AVCodec_find_decoder(stream.CodecPar().CodecID())
	if codec == nil {
		ff.AVUtil_frame_free(dec.frame)
		return nil, errors.New("failed to find decoder for codec")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		ff.AVUtil_frame_free(dec.frame)
		return nil, errors.New("failed to allocate codec context")
	}
	dec.codec = ctx

	// Copy codec parameters from stream to codec context
	if err := ff.AVCodec_parameters_to_context(dec.codec, stream.CodecPar()); err != nil {
		dec.close()
		return nil, err
	}

	// Open codec
	if err := ff.AVCodec_open(dec.codec, codec, nil); err != nil {
		dec.close()
		return nil, err
	}

	// Create resampler if destination parameters differ from source or force is set
	if destPar != nil && (force || parametersNeedResampling(srcPar, destPar)) {
		resampler, err := NewResampler(destPar, force)
		if err != nil {
			dec.close()
			return nil, err
		}
		dec.resampler = resampler
	}

	return dec, nil
}

// Check if two parameter sets require resampling/rescaling
func parametersNeedResampling(src, dest *Par) bool {
	if src == nil || dest == nil {
		return false
	}

	// Must be same type to compare
	if src.Type() != dest.Type() {
		return true
	}

	// Check based on type
	switch src.Type() {
	case media.AUDIO:
		// Audio needs resampling if format, rate, or channels differ
		if src.SampleFormat() != dest.SampleFormat() {
			return true
		}
		if src.SampleRate() != dest.SampleRate() {
			return true
		}
		srcCh := src.ChannelLayout()
		destCh := dest.ChannelLayout()
		return !ff.AVUtil_channel_layout_compare(&srcCh, &destCh)
	case media.VIDEO:
		// Video needs rescaling if format, width, or height differ
		if src.PixelFormat() != dest.PixelFormat() {
			return true
		}
		if src.Width() != dest.Width() || src.Height() != dest.Height() {
			return true
		}
		return false
	default:
		return false
	}
}

// Close and free stream decoder resources
func (dec *streamDecoder) close() {
	if dec.resampler != nil {
		dec.resampler.Close()
		dec.resampler = nil
	}
	if dec.codec != nil {
		ff.AVCodec_free_context(dec.codec)
		dec.codec = nil
	}
	if dec.frame != nil {
		ff.AVUtil_frame_free(dec.frame)
		dec.frame = nil
	}
}

// Decode a packet and emit frames or subtitles via the callbacks
func (dec *streamDecoder) decode(pkt *ff.AVPacket, framefn DecoderFrameFn, subtitlefn DecoderSubtitleFn) error {
	// Handle subtitle decoding separately (uses legacy API)
	if dec.codec.CodecType() == ff.AVMEDIA_TYPE_SUBTITLE {
		return dec.decodeSubtitle(pkt, subtitlefn)
	}

	// Set packet timebase
	if pkt != nil {
		pkt.SetTimeBase(dec.timeBase)
	}

	// Send packet to decoder (nil packet flushes)
	if err := ff.AVCodec_send_packet(dec.codec, pkt); err != nil {
		return err
	}

	// Receive all available frames
	for {
		// Receive frame from decoder
		err := ff.AVCodec_receive_frame(dec.codec, dec.frame)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, syscall.EAGAIN) {
				// No more frames available
				break
			}
			return err
		}

		// Set frame timebase
		dec.frame.SetTimeBase(dec.timeBase)

		// Resample/rescale if needed
		srcFrame := (*Frame)(dec.frame)
		if dec.resampler != nil {
			if err := dec.resampler.Resample(srcFrame, func(destFrame *Frame) error {
				if destFrame != nil {
					return framefn(dec.stream, destFrame)
				}
				return nil
			}); err != nil {
				return err
			}
		} else {
			// No resampling, pass frame directly
			if err := framefn(dec.stream, srcFrame); err != nil {
				return err
			}
		}

		// Unref the frame for next iteration
		ff.AVUtil_frame_unref(dec.frame)
	}

	return nil
}

// Decode a subtitle packet using the legacy subtitle API
func (dec *streamDecoder) decodeSubtitle(pkt *ff.AVPacket, subtitlefn DecoderSubtitleFn) error {
	// Subtitles don't support flushing (nil packet)
	if pkt == nil {
		return nil
	}

	// Skip if no callback provided
	if subtitlefn == nil {
		return nil
	}

	// Set packet timebase
	pkt.SetTimeBase(dec.timeBase)

	// Decode subtitle using legacy API
	sub, err := ff.AVCodec_decode_subtitle(dec.codec, pkt)
	if err != nil {
		return err
	}
	if sub == nil {
		// No subtitle in this packet
		return nil
	}
	defer ff.AVSubtitle_free(sub)

	// Call user callback with decoded subtitle
	return subtitlefn(dec.stream, sub)
}
