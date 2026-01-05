package ffmpeg

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Media reader which reads from a URL, file path or device
type Reader struct {
	t       media.Type
	input   *ff.AVFormatContext
	avio    *ff.AVIOContextEx
	force   bool
	context *Context
}

type reader_callback struct {
	r io.Reader
}

// Return parameters if a stream should be decoded and either resampled or
// resized. Return nil if you want to ignore the stream, or pass back the
// stream parameters if you want to copy the stream without any changes.
type DecoderMapFunc func(int, *Par) (*Par, error)

// DecoderFrameFn is a function which is called to send a frame after decoding. It should
// return nil to continue decoding or io.EOF to stop.
type DecoderFrameFn func(int, *Frame) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open media from a url, file path or device
func Open(url string, opt ...Opt) (*Reader, error) {
	options := newOpts()
	reader := new(Reader)

	// Apply options
	for _, opt := range opt {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	// Get the options
	dict := ff.AVUtil_dict_alloc()
	defer ff.AVUtil_dict_free(dict)
	if len(options.opts) > 0 {
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(options.opts, " "), "=", " ", 0); err != nil {
			return nil, err
		}
	}

	// Open the device or stream
	if ctx, err := ff.AVFormat_open_url(url, options.iformat, dict); err != nil {
		return nil, err
	} else {
		reader.input = ctx
	}

	// Find stream information and do rest of the initialization
	return reader.open(options)
}

// Create a new reader from an io.Reader
func NewReader(r io.Reader, opt ...Opt) (*Reader, error) {
	options := newOpts()
	reader := new(Reader)

	// Apply options
	for _, opt := range opt {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	// Get the options
	dict := ff.AVUtil_dict_alloc()
	defer ff.AVUtil_dict_free(dict)
	if len(options.opts) > 0 {
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(options.opts, " "), "=", " ", 0); err != nil {
			return nil, err
		}
	}

	// Allocate the AVIO context
	reader.avio = ff.AVFormat_avio_alloc_context(bufSize, false, &reader_callback{r})
	if reader.avio == nil {
		return nil, errors.New("failed to allocate avio context")
	}

	// Open the stream
	if ctx, err := ff.AVFormat_open_reader(reader.avio, options.iformat, dict); err != nil {
		ff.AVFormat_avio_context_free(reader.avio)
		return nil, err
	} else {
		reader.input = ctx
	}

	// Find stream information and do rest of the initialization
	return reader.open(options)
}

func (r *Reader) open(options *opts) (*Reader, error) {
	// Find stream information
	if err := ff.AVFormat_find_stream_info(r.input, nil); err != nil {
		ff.AVFormat_free_context(r.input)
		return nil, err
	}

	// Set force flag and type
	r.force = options.force
	r.t = options.t | media.INPUT

	// Return success
	return r, nil
}

// Close the reader
func (r *Reader) Close() error {
	var result error

	// Free context
	if ctx := r.context; ctx != nil {
		result = errors.Join(result, ctx.Close())
	}

	// Free resources
	ff.AVFormat_free_context(r.input)
	if r.avio != nil {
		ff.AVFormat_avio_context_free(r.avio)
	}

	// Release resources
	r.context = nil
	r.input = nil
	r.avio = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Display the reader as a string
func (r *Reader) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.input)
}

// Display the reader as a string
func (r *Reader) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the media type
func (r *Reader) Type() media.Type {
	return r.t
}

// Return the duration of the media stream, returns zero if unknown
func (r *Reader) Duration() time.Duration {
	duration := r.input.Duration()
	if duration > 0 {
		return time.Duration(duration) * time.Second / time.Duration(ff.AV_TIME_BASE)
	}
	return 0
}

// Return the "best stream" for a specific media type, or -1 if there is no
// "best stream" for that type.
func (r *Reader) BestStream(t media.Type) int {
	switch {
	case t.Is(media.VIDEO):
		if stream, _, err := ff.AVFormat_find_best_stream(r.input, ff.AVMEDIA_TYPE_VIDEO, -1, -1); err == nil {
			// Only return if this doesn't have a disposition - so we don't select artwork, for example
			disposition := r.input.Stream(stream).Disposition()
			if disposition == 0 || disposition.Is(ff.AV_DISPOSITION_DEFAULT) {
				return r.input.Stream(stream).Index()
			}
		}
	case t.Is(media.AUDIO):
		if stream, _, err := ff.AVFormat_find_best_stream(r.input, ff.AVMEDIA_TYPE_AUDIO, -1, -1); err == nil {
			return r.input.Stream(stream).Index()
		}
	case t.Is(media.SUBTITLE):
		if stream, _, err := ff.AVFormat_find_best_stream(r.input, ff.AVMEDIA_TYPE_SUBTITLE, -1, -1); err == nil {
			return r.input.Stream(stream).Index()
		}
	case t.Is(media.DATA):
		if stream, _, err := ff.AVFormat_find_best_stream(r.input, ff.AVMEDIA_TYPE_DATA, -1, -1); err == nil {
			return r.input.Stream(stream).Index()
		}
	}
	return -1
}

// Return all streams of a specific type (video, audio, subtitle, data)
func (r *Reader) Streams(t media.Type) []*Stream {
	var result []*Stream
	for _, stream := range r.input.Streams() {
		switch stream.CodecPar().CodecType() {
		case ff.AVMEDIA_TYPE_VIDEO:
			if t.Is(media.VIDEO) || t == media.ANY {
				result = append(result, newStream(stream))
			}
		case ff.AVMEDIA_TYPE_AUDIO:
			if t.Is(media.AUDIO) || t == media.ANY {
				result = append(result, newStream(stream))
			}
		case ff.AVMEDIA_TYPE_SUBTITLE:
			if t.Is(media.SUBTITLE) || t == media.ANY {
				result = append(result, newStream(stream))
			}
		case ff.AVMEDIA_TYPE_DATA:
			if t.Is(media.DATA) || t == media.ANY {
				result = append(result, newStream(stream))
			}
		case ff.AVMEDIA_TYPE_ATTACHMENT:
			if t.Is(media.DATA) || t == media.ANY {
				result = append(result, newStream(stream))
			}
		}
	}

	// Return the streams
	return result
}

// Return the metadata for the media stream, filtering by the specified keys
// if there are any. Artwork is returned with the "artwork" key.
func (r *Reader) Metadata(keys ...string) []*Metadata {
	entries := ff.AVUtil_dict_entries(r.input.Metadata())
	result := make([]*Metadata, 0, len(entries))
	for _, entry := range entries {
		if len(keys) == 0 || slices.Contains(keys, entry.Key()) {
			result = append(result, NewMetadata(entry.Key(), entry.Value()))
		}
	}

	// Obtain any artwork from the streams
	if slices.Contains(keys, MetaArtwork) {
		for _, stream := range r.input.Streams() {
			if packet := stream.AttachedPic(); packet != nil {
				result = append(result, NewMetadata(MetaArtwork, packet.Bytes()))
			}
		}
	}

	// Return all the metadata
	return result
}

// Seek to a specific time in the media stream, in seconds
func (r *Reader) Seek(stream int, secs float64) error {
	ctx := r.input.Stream(stream)
	if ctx == nil {
		return media.ErrBadParameter.With("stream not found")
	}
	// At the moment, it seeks to the previous keyframe
	tb := int64(secs / ff.AVUtil_rational_q2d(ctx.TimeBase()))
	return ff.AVFormat_seek_frame(r.input, ctx.Index(), tb, ff.AVSEEK_FLAG_BACKWARD)
}

func (r *Reader) Decode(ctx context.Context, mapfn DecoderMapFunc, decodefn DecoderFrameFn) error {
	// Create a decoding context
	decoders, err := newContext(r, mapfn)
	if err != nil {
		return err
	}
	defer decoders.Close()

	// Do the decoding
	return decoders.decode(ctx, decodefn)
}

// Map streams to decoders, and return the decoding context
// The map function is called for each stream
// and should return the parameters for the destination frame. If any
// parameters are returned as null, then that stream is ignored
func (r *Reader) Map(fn DecoderMapFunc) (*Context, error) {
	return newContext(r, fn)
}

// Decode the media stream into frames. The decodefn is called for each
// frame decoded from the stream.
//
// The decoding can be interrupted by cancelling the context, or by the decodefn
// returning an error or io.EOF. The latter will end the decoding process early but
// will not return an error.
func (r *Reader) DecodeWithContext(ctx context.Context, decoders *Context, decodefn DecoderFrameFn) error {
	return decoders.decode(ctx, decodefn)
}

// Transcode the media stream to a writer
// As per the decode method, the map function is called for each stream and should return the
// parameters for the destination. If the map function returns nil for a stream, then
// the stream is ignored.
/*
func (r *Reader) Transcode(ctx context.Context, w io.Writer, mapfn DecoderMapFunc, opt ...Opt) error {
	// Map streams to decoders
	decoders, err := r.mapStreams(mapfn)
	if err != nil {
		return err
	}
	defer decoders.Close()

	// Add streams to the output
	for _, decoder := range decoders {
		opt = append(opt, OptStream(decoder.stream, decoder.par))
	}

	// Create an output
	output, err := NewWriter(w, opt...)
	if err != nil {
		return err
	}
	defer output.Close()

	// One go-routine for decoding, one for encoding
	var wg sync.WaitGroup
	var result error

	// Make a channel for transcoding frames. The decoder should
	// be ahead of the encoder, so there is probably no need to
	// create a buffered channel.
	ch := make(chan *Frame)

	// Decoding
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := r.decode(ctx, decoders, func(stream int, frame *Frame) error {
			ch <- frame
			return nil
		}); err != nil {
			result = err
		}
		// Close channel at the end of decoding
		close(ch)
	}()

	// Encoding
	wg.Add(1)
	go func() {
		defer wg.Done()
		for frame := range ch {
			fmt.Println("TODO: Write frame to output", frame)
		}
	}()

	// Wait for the process to finish
	wg.Wait()

	// Return any errors
	return result
}
*/

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - CALLBACK

func (r *reader_callback) Reader(buf []byte) int {
	n, err := r.r.Read(buf)
	if err != nil {
		return ff.AVERROR_EOF
	}
	return n
}

func (r *reader_callback) Seeker(offset int64, whence int) int64 {
	whence = whence & ^ff.AVSEEK_FORCE
	seeker, ok := r.r.(io.ReadSeeker)
	if !ok {
		return -1
	}
	switch whence {
	case io.SeekStart, io.SeekCurrent, io.SeekEnd:
		n, err := seeker.Seek(offset, whence)
		if err != nil {
			return -1
		}
		return n
	}
	return -1
}

func (r *reader_callback) Writer([]byte) int {
	return ff.AVERROR_EOF
}
