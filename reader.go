package media

import (
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type reader struct {
	t       MediaType
	input   *ff.AVFormatContext
	avio    *ff.AVIOContextEx
	demuxer *demuxer
	force   bool // passed my the manager object
}

type reader_callback struct {
	r io.Reader
}

var _ Media = (*reader)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bufSize = 4096
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open media from a url, file path or device
func newMedia(url string, format Format, opts ...string) (*reader, error) {
	reader := new(reader)
	reader.t = INPUT

	// Set the input format
	var f *ff.AVInputFormat
	if format != nil {
		reader.t |= format.Type()
		if inputfmt, ok := format.(*inputformat); ok {
			f = inputfmt.ctx
		}
	}

	// Get the options
	dict := ff.AVUtil_dict_alloc()
	defer ff.AVUtil_dict_free(dict)
	if len(opts) > 0 {
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(opts, " "), "=", " ", 0); err != nil {
			return nil, err
		}
	}

	// Open the device or stream
	if ctx, err := ff.AVFormat_open_url(url, f, dict); err != nil {
		return nil, err
	} else {
		reader.input = ctx
	}

	// Find stream information and do rest of the initialization
	return reader.open()
}

// Create a new reader from an io.Reader
func newReader(r io.Reader, format Format, opts ...string) (*reader, error) {
	reader := new(reader)
	reader.t = INPUT | FILE

	// Set the input format
	var fmt *ff.AVInputFormat
	if format != nil {
		reader.t |= format.Type()
		if format.Type().Is(DEVICE) {
			return nil, errors.New("cannot create a reader from a device")
		}
		if inputfmt, ok := format.(*inputformat); ok {
			fmt = inputfmt.ctx
		}
	}

	// Get the options
	dict := ff.AVUtil_dict_alloc()
	defer ff.AVUtil_dict_free(dict)
	if len(opts) > 0 {
		if err := ff.AVUtil_dict_parse_string(dict, strings.Join(opts, " "), "=", " ", 0); err != nil {
			return nil, err
		}
	}

	// Allocate the AVIO context
	reader.avio = ff.AVFormat_avio_alloc_context(bufSize, false, &reader_callback{r})
	if reader.avio == nil {
		return nil, errors.New("failed to allocate avio context")
	}

	// Open the stream
	if ctx, err := ff.AVFormat_open_reader(reader.avio, fmt, dict); err != nil {
		ff.AVFormat_avio_context_free(reader.avio)
		return nil, err
	} else {
		reader.input = ctx
	}

	// Find stream information and do rest of the initialization
	return reader.open()
}

func (r *reader) open() (*reader, error) {
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
func (r *reader) Close() error {
	var result error

	// Free demuxer
	if r.demuxer != nil {
		result = errors.Join(result, r.demuxer.close())
	}

	// Free resources
	ff.AVFormat_free_context(r.input)
	if r.avio != nil {
		ff.AVFormat_avio_context_free(r.avio)
	}

	// Release resources
	r.demuxer = nil
	r.input = nil
	r.avio = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Display the reader as a string
func (r *reader) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.input)
}

// Display the reader as a string
func (r *reader) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (r *reader) Type() MediaType {
	return r.t
}

func (r *reader) Decoder(fn DecoderMapFunc) (Decoder, error) {
	// Check if this is actually an input
	if !r.Type().Is(INPUT) {
		return nil, errors.New("not an input stream")
	}

	// Return existing decoder
	if r.demuxer != nil {
		return r.demuxer, nil
	}

	// Create a decoding context
	decoder, err := newDemuxer(r.input, fn, r.force)
	if err != nil {
		return nil, err
	} else {
		r.demuxer = decoder
	}

	// Return success
	return decoder, nil
}

// Return the metadata for the media stream, filtering
// by the specified keys if there are any. Artwork
// is returned by using the "artwork" key.
func (r *reader) Metadata(keys ...string) []Metadata {
	entries := ff.AVUtil_dict_entries(r.input.Metadata())
	result := make([]Metadata, 0, len(entries))
	for _, entry := range entries {
		if len(keys) == 0 || slices.Contains(keys, entry.Key()) {
			result = append(result, newMetadata(entry.Key(), entry.Value()))
		}
	}

	// Obtain any artwork from the streams
	if slices.Contains(keys, MetaArtwork) {
		for _, stream := range r.input.Streams() {
			if packet := stream.AttachedPic(); packet != nil {
				result = append(result, newMetadata(MetaArtwork, packet.Bytes()))
			}
		}
	}

	// Return all the metadata
	return result
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
