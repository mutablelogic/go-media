package reader

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Reader is a wrapper around an AVFormatContext that provides a higher-level
// interface for reading media files.
type Reader struct {
	sync.Mutex
	opts
	input *ff.AVFormatContext
	avio  *ff.AVIOContextEx
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open media from a url, file path or device
func Open(url string, opt ...Opt) (*Reader, error) {
	self := new(Reader)

	// Set reader options
	if err := self.opts.apply(opt...); err != nil {
		return nil, err
	}

	// Get the options
	dict := ff.AVUtil_dict_alloc()
	defer ff.AVUtil_dict_free(dict)
	if len(self.options) > 0 {
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(self.options, " "), "=", " ", ff.AV_DICT_NONE); err != nil {
			return nil, err
		}
	}

	// Open the device or stream
	if input, err := ff.AVFormat_open_url(url, self.opts.input, dict); err != nil {
		return nil, err
	} else {
		self.input = input
	}

	// Find stream information and do rest of the initialization
	return self.open()
}

// Create a new reader from an io.Reader
func NewReader(r io.Reader, opt ...Opt) (*Reader, error) {
	self := new(Reader)

	// Set reader options
	if err := self.opts.apply(opt...); err != nil {
		return nil, err
	}

	// Get the options
	dict := ff.AVUtil_dict_alloc()
	defer ff.AVUtil_dict_free(dict)
	if len(self.options) > 0 {
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(self.options, " "), "=", " ", ff.AV_DICT_NONE); err != nil {
			return nil, err
		}
	}

	// Allocate the AVIO context
	self.avio = ff.AVFormat_avio_alloc_context(ioBufferSize, false, &io_callback{r})
	if self.avio == nil {
		return nil, gomedia.ErrInternalError.Withf("failed to allocate AVIO context")
	}

	// Open the stream
	if ctx, err := ff.AVFormat_open_reader(self.avio, self.opts.input, dict); err != nil {
		ff.AVFormat_avio_context_free(self.avio)
		return nil, err
	} else {
		self.input = ctx
	}

	// Find stream information and do rest of the initialization
	return self.open()
}

func (r *Reader) open() (*Reader, error) {
	// Find stream information
	if err := ff.AVFormat_find_stream_info(r.input, nil); err != nil {
		ff.AVFormat_free_context(r.input)
		r.input = nil
		if r.avio != nil {
			ff.AVFormat_avio_context_free(r.avio)
			r.avio = nil
		}
		return nil, err
	}

	// Return success
	return r, nil
}

// Close the reader
func (r *Reader) Close() error {
	var result error

	// Mutex lock to ensure thread safety
	r.Lock()
	defer r.Unlock()

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
// PUBLIC METHODS

// Return the duration of the media stream, returns zero if unknown
func (r *Reader) Duration() time.Duration {
	if r.input == nil {
		return 0
	}
	duration := r.input.Duration()
	if duration > 0 {
		// Divide before multiplying: duration is in AV_TIME_BASE (µs) units,
		// and multiplying by time.Second first overflows int64 for anything
		// longer than ~2.56 hours. time.Second/AV_TIME_BASE is exact (1e9/1e6
		// = 1000), so reordering loses no precision.
		return time.Duration(duration) * (time.Second / time.Duration(ff.AV_TIME_BASE))
	}
	return 0
}

// Seek to a specific time in the media stream
func (r *Reader) Seek(stream int, d time.Duration) error {
	if r.input == nil {
		return gomedia.ErrInternalError.With("reader is closed")
	}
	ctx := r.input.Stream(stream)
	if ctx == nil {
		return gomedia.ErrBadParameter.With("stream not found")
	}
	// Rescale from nanoseconds (time.Duration's unit) to the stream's own
	// timebase using exact integer arithmetic, rather than a float64
	// division which would lose precision for large durations.
	tb := ff.AVUtil_rational_rescale_q(int64(d), ff.AVUtil_rational(1, int(time.Second)), ctx.TimeBase())
	// At the moment, it seeks to the previous keyframe
	return ff.AVFormat_seek_frame(r.input, ctx.Index(), tb, ff.AVSEEK_FLAG_BACKWARD)
}

// Return the metadata for the media stream, filtering by the specified keys
// if there are any. Artwork is returned with the "artwork" key.
func (r *Reader) Metadata(keys ...string) []gomedia.Metadata {
	if r.input == nil {
		return nil
	}

	entries := ff.AVUtil_dict_entries(r.input.Metadata())
	result := make([]gomedia.Metadata, 0, len(entries))
	for _, entry := range entries {
		// Artwork is handled separately below, from attached-pic streams
		if entry.Key() == gomedia.MetaArtwork {
			continue
		}
		if len(keys) == 0 || slices.Contains(keys, entry.Key()) {
			result = append(result, &meta{key: entry.Key(), value: entry.Value()})
		}
	}

	// Obtain any artwork from the streams
	if len(keys) == 0 || slices.Contains(keys, gomedia.MetaArtwork) {
		for _, stream := range r.input.Streams() {
			if packet := stream.AttachedPic(); packet != nil {
				result = append(result, &meta{key: gomedia.MetaArtwork, value: packet.Bytes()})
			}
		}
	}

	// Return all the metadata
	return result
}

// Return the audio, video, and subtitle streams in the media file as
// profiles, keyed by their real stream index — mirroring the shape
// writer.WithProfile expects, so a caller can remux directly:
//
//	for i, p := range r.Streams() {
//	    opts = append(opts, writer.WithProfile(i, p))
//	}
//
// Data/attachment streams and attached-pic (cover art) streams are omitted;
// artwork is available via Metadata's "artwork" key instead.
func (r *Reader) Streams() map[int]profile.Profile {
	if r.input == nil {
		return nil
	}

	result := make(map[int]profile.Profile)
	for _, stream := range r.input.Streams() {
		sp, err := profile.NewStreamProfile(stream)
		if err != nil {
			continue
		}
		result[sp.Index()] = sp
	}
	return result
}

// Decode packets from the media stream without decoding to frames. The packetfn is called for each
// packet read from any stream. Use this for stream copying or remuxing without transcoding.
//
// The reading can be interrupted by cancelling the context, or by the packetfn
// returning an error or io.EOF. The latter will end the reading process early but
// will not return an error.
func (r *Reader) Decode(ctx context.Context, packetfn PacketFn) error {
	// Lock decoding
	r.Lock()
	defer r.Unlock()

	if r.input == nil {
		return gomedia.ErrInternalError.With("reader is closed")
	}

	// Allocate packet for reading - this will be reused for each read
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		return gomedia.ErrInternalError.With("failed to allocate packet")
	}
	defer ff.AVCodec_packet_free(packet)

	// Read packets until EOF, context cancellation or error
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Unref any previous packet data before reading
		ff.AVCodec_packet_unref(packet)

		// Read next packet from any stream
		if err := ff.AVFormat_read_frame(r.input, packet); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if errors.Is(err, syscall.EAGAIN) {
				// No data available right now - retry
				continue
			}
			return err
		}

		if err := packetfn(packet); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
