package reader

import (
	"errors"
	"io"
	"sync"
	"syscall"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// PacketFn is called for each packet read by Reader.Decode. Returning io.EOF
// stops decoding early without being treated as an error.
type PacketFn func(*ff.AVPacket) error

// FrameFn is called for each frame a Decoder produces - an *frame.AudioFrame
// or *frame.VideoFrame for audio/video streams, or a *frame.SubtitleFrame
// for subtitle streams (which FFmpeg decodes via a completely separate,
// legacy API). Returning io.EOF stops decoding early without being treated
// as an error.
type FrameFn func(frame.Frame) error

// Decoder decodes packets from a set of registered streams into frames,
// passing each frame to fn. Pass a Decoder to Reader.Decode alongside a
// PacketFn to decode selected streams while still seeing every packet.
//
// Streams are opted in with Add. A stream that isn't registered is
// discarded from the decoder's point of view: Reader.Decode still passes
// its packets to the PacketFn (if any), but they are never sent to a codec.
type Decoder struct {
	sync.RWMutex
	fn      FrameFn
	include map[int]bool
	codecs  map[int]*ff.AVCodecContext
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewDecoder creates an empty Decoder that passes every frame it decodes to
// fn. Use Add to register which streams should be decoded before passing
// the Decoder to Reader.Decode.
func NewDecoder(fn FrameFn) (*Decoder, error) {
	if fn == nil {
		return nil, gomedia.ErrBadParameter.With("nil callback function")
	}
	return &Decoder{fn: fn, include: make(map[int]bool)}, nil
}

// Add registers stream for decoding. Streams which are never registered
// are left alone by Reader.Decode - this is how you discard a stream from
// frame decoding without affecting the PacketFn.
func (d *Decoder) Add(stream int) error {
	d.Lock()
	defer d.Unlock()

	if d.include[stream] {
		return gomedia.ErrBadParameter.Withf("stream %d already registered", stream)
	}
	d.include[stream] = true
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - used by Reader.Decode

// open allocates and opens a codec context for each stream registered with
// Add, using that stream's own codec parameters and timebase.
func (d *Decoder) open(input *ff.AVFormatContext) error {
	d.Lock()
	defer d.Unlock()

	d.codecs = make(map[int]*ff.AVCodecContext, len(d.include))
	for _, stream := range input.Streams() {
		index := stream.Index()
		if !d.include[index] {
			continue
		}

		par := stream.CodecPar()
		codec := ff.AVCodec_find_decoder(par.CodecID())
		if codec == nil {
			return gomedia.ErrBadParameter.Withf("no decoder found for stream %d", index)
		}

		ctx := ff.AVCodec_alloc_context(codec)
		if ctx == nil {
			return gomedia.ErrInternalError.With("failed to allocate codec context")
		}

		if err := ff.AVCodec_parameters_to_context(ctx, par); err != nil {
			ff.AVCodec_free_context(ctx)
			return err
		}
		ctx.SetTimeBase(stream.TimeBase())

		if err := ff.AVCodec_open(ctx, codec, nil); err != nil {
			ff.AVCodec_free_context(ctx)
			return err
		}

		d.codecs[index] = ctx
	}
	return nil
}

// close releases every codec context opened by open.
func (d *Decoder) close() error {
	d.Lock()
	defer d.Unlock()

	for _, ctx := range d.codecs {
		ff.AVCodec_free_context(ctx)
	}
	d.codecs = nil
	return nil
}

// decode sends packet (or nil, to flush) to the codec registered for
// stream and passes any resulting frames to fn. If stream has no
// registered codec, this is a no-op - the stream was discarded via Add.
func (d *Decoder) decode(stream int, packet *ff.AVPacket) error {
	d.RLock()
	ctx, ok := d.codecs[stream]
	d.RUnlock()
	if !ok {
		return nil
	}

	if ctx.CodecType() == ff.AVMEDIA_TYPE_SUBTITLE {
		return d.decodeSubtitle(ctx, stream, packet)
	}
	return d.decodeFrame(ctx, stream, packet)
}

// decodeFrame handles the audio/video path: send_packet/receive_frame.
func (d *Decoder) decodeFrame(ctx *ff.AVCodecContext, stream int, packet *ff.AVPacket) error {
	if err := ff.AVCodec_send_packet(ctx, packet); err != nil {
		return err
	}

	isVideo := ctx.CodecType() == ff.AVMEDIA_TYPE_VIDEO

	for {
		var f frame.Frame
		var raw *ff.AVFrame

		if isVideo {
			vf, err := frame.NewVideoFrame(stream)
			if err != nil {
				return err
			}
			f, raw = vf, vf.AVFrame
		} else {
			af, err := frame.NewAudioFrame(stream)
			if err != nil {
				return err
			}
			f, raw = af, af.AVFrame
		}

		if err := ff.AVCodec_receive_frame(ctx, raw); err != nil {
			f.Close()
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		raw.SetTimeBase(ctx.TimeBase())

		err := d.fn(f)
		f.Close()
		if err != nil {
			return err
		}
	}
}

// decodeSubtitle handles the subtitle path: FFmpeg decodes subtitles via a
// separate, legacy, non-streaming API (one packet in, at most one subtitle
// out - no buffering, so unlike decodeFrame there is nothing to flush).
func (d *Decoder) decodeSubtitle(ctx *ff.AVCodecContext, stream int, packet *ff.AVPacket) error {
	if packet == nil {
		// Subtitles don't support flushing
		return nil
	}

	sub, err := ff.AVCodec_decode_subtitle(ctx, packet)
	if err != nil {
		return err
	}
	if sub == nil {
		// No subtitle produced from this packet
		return nil
	}

	f := frame.NewSubtitleFrame(stream, sub)
	err = d.fn(f)
	f.Close()
	return err
}

// flush drains any frames buffered by every registered codec, at end of
// stream.
func (d *Decoder) flush() error {
	d.RLock()
	streams := make([]int, 0, len(d.codecs))
	for stream := range d.codecs {
		streams = append(streams, stream)
	}
	d.RUnlock()

	for _, stream := range streams {
		if err := d.decode(stream, nil); err != nil {
			return err
		}
	}
	return nil
}
