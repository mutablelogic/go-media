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
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Media reader which reads from a URL, file path or device
type Reader struct {
	input *ff.AVFormatContext
	avio  *ff.AVIOContextEx
}

type reader_callback struct {
	r io.Reader
}

// Return parameters if a stream should be decoded and either resampled or
// resized. Return nil if you want to ignore the stream, or pass back the
// stream parameters if you want to copy the stream without any changes.
type DecoderMapFunc func(int, *Par) (*Par, error)

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

func (r *Reader) open(_ *opts) (*Reader, error) {
	// Find stream information
	if err := ff.AVFormat_find_stream_info(r.input, nil); err != nil {
		ff.AVFormat_free_context(r.input)
		ff.AVFormat_avio_context_free(r.avio)
		return nil, err
	}

	// Return success
	return r, nil
}

// Close the reader
func (r *Reader) Close() error {
	var result error

	// Free resources
	ff.AVFormat_free_context(r.input)
	if r.avio != nil {
		ff.AVFormat_avio_context_free(r.avio)
	}

	// Release resources
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

// Return the duration of the media stream, returns zero if unknown
func (r *Reader) Duration() time.Duration {
	duration := r.input.Duration()
	if duration > 0 {
		return time.Duration(duration) * time.Second / time.Duration(ff.AV_TIME_BASE)
	}
	return 0
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

// TODO Decode the media stream into packets and frames
func (r *Reader) Decode(ctx context.Context, fn DecoderMapFunc) error {
	return errors.New("not implemented yet")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

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
