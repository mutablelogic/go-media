package ffmpeg

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Media reader which reads from a URL, file path or device
type Reader struct {
	mu    sync.Mutex
	t     media.Type
	input *ff.AVFormatContext
	avio  *ff.AVIOContextEx
	force bool
}

type reader_callback struct {
	r io.Reader
}

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
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(options.opts, " "), "=", " ", ff.AV_DICT_NONE); err != nil {
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
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(options.opts, " "), "=", " ", ff.AV_DICT_NONE); err != nil {
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
	r.mu.Lock()
	defer r.mu.Unlock()

	var result error

	// Free resources
	if r.input != nil {
		ff.AVFormat_free_context(r.input)
		r.input = nil
	}
	if r.avio != nil {
		ff.AVFormat_avio_context_free(r.avio)
		r.avio = nil
	}

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

// Return the input format (for probing format info)
func (r *Reader) InputFormat() *ff.AVInputFormat {
	return r.input.Input()
}

// Return the duration of the media stream, returns zero if unknown
func (r *Reader) Duration() time.Duration {
	duration := r.input.Duration()
	if duration > 0 {
		return time.Duration(duration) * time.Second / time.Duration(ff.AV_TIME_BASE)
	}
	return 0
}

// Return the raw AVStream objects for direct access
func (r *Reader) AVStreams() []*ff.AVStream {
	return r.input.Streams()
}

// Return all streams of a specific type (video, audio, subtitle, data)
// Use media.ANY to return all streams regardless of type
func (r *Reader) Streams(t media.Type) []*schema.Stream {
	streams := r.input.Streams()
	result := make([]*schema.Stream, 0, len(streams))

	// If ANY is requested, return all streams
	if t == media.ANY {
		for _, stream := range streams {
			if s := schema.NewStream(stream); s != nil {
				result = append(result, s)
			}
		}
		return result
	}

	// Otherwise filter by type using Stream.Type() which handles all the mapping
	for _, stream := range streams {
		s := schema.NewStream(stream)
		if s != nil && s.Type().Is(t) {
			result = append(result, s)
		}
	}

	return result
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

// Decode packets from the media stream without decoding to frames. The packetfn is called for each
// packet read from any stream. Use this for stream copying or remuxing without transcoding.
//
// The reading can be interrupted by cancelling the context, or by the packetfn
// returning an error or io.EOF. The latter will end the reading process early but
// will not return an error.
func (r *Reader) Decode(ctx context.Context, packetfn DecoderPacketFn) error {
	// Check reader is valid
	r.mu.Lock()
	if r.input == nil {
		r.mu.Unlock()
		return errors.New("reader is closed")
	}
	r.mu.Unlock()

	// Create decoder
	dec, err := newDecoder(r)
	if err != nil {
		return err
	}
	defer dec.free()

	// Read packets
	return dec.readPackets(ctx, packetfn)
}

// Demux and decode the media stream into frames. The map function determines which
// streams to decode and what output parameters to use. The framefn is called for each
// decoded frame from any mapped stream.
//
// The decoding can be interrupted by cancelling the context, or by the framefn
// returning an error or io.EOF. The latter will end the decoding process early but
// will not return an error.
func (r *Reader) Demux(ctx context.Context, mapfn DecoderMapFunc, framefn DecoderFrameFn) error {
	// Check reader is valid
	r.mu.Lock()
	if r.input == nil {
		r.mu.Unlock()
		return errors.New("reader is closed")
	}
	r.mu.Unlock()

	// Create decoder
	dec, err := newDecoder(r)
	if err != nil {
		return err
	}
	defer dec.free()

	// Decode frames
	return dec.decodeFrames(ctx, mapfn, framefn)
}

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
