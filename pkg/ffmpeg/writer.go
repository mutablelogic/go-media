package ffmpeg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
	maps "golang.org/x/exp/maps"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Create media from io.Writer
type Writer struct {
	output   *ff.AVFormatContext
	header   bool
	encoders []*Encoder
}

type writer_callback struct {
	w io.Writer
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bufSize = 4096
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new writer with a URL and options
func Create(url string, opt ...Opt) (*Writer, error) {
	options := newOpts()
	writer := new(Writer)

	// Apply options
	for _, opt := range opt {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	// Guess the output format
	var ofmt *ff.AVOutputFormat
	if options.oformat == nil && url != "" {
		options.oformat = ff.AVFormat_guess_format("", url, "")
	}
	if options.oformat == nil {
		return nil, ErrBadParameter.With("unable to guess the output format")
	}

	// Allocate the output media context
	ctx, err := ff.AVFormat_create_file(url, ofmt)
	if err != nil {
		return nil, err
	} else {
		writer.output = ctx
	}

	// Continue with open
	return writer.open(options)
}

// Create a new writer with an io.Writer and options
func NewWriter(w io.Writer, opt ...Opt) (*Writer, error) {
	options := newOpts()
	writer := new(Writer)

	// Apply options
	for _, opt := range opt {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	// Check output
	var filename string
	if options.oformat == nil {
		return nil, ErrBadParameter.Withf("invalid output format")
	} else if w_, ok := w.(*os.File); ok {
		filename = w_.Name()
	}

	// Allocate the AVIO context
	avio := ff.AVFormat_avio_alloc_context(bufSize, true, &writer_callback{w})
	if avio == nil {
		return nil, errors.New("failed to allocate avio context")
	} else if ctx, err := ff.AVFormat_open_writer(avio, options.oformat, filename); err != nil {
		return nil, err
	} else {
		writer.output = ctx
	}

	// Continue with open
	return writer.open(options)
}

func (writer *Writer) open(options *opts) (*Writer, error) {
	// Create codec contexts for each stream
	var result error
	keys := sort.IntSlice(maps.Keys(options.streams))
	for _, stream := range keys {
		encoder, err := NewEncoder(writer.output, stream, options.streams[stream])
		if err != nil {
			result = errors.Join(result, err)
			continue
		} else {
			writer.encoders = append(writer.encoders, encoder)
		}
	}

	// Return any errors
	if result != nil {
		return nil, errors.Join(result, writer.Close())
	}

	// Add metadata
	metadata := ff.AVUtil_dict_alloc()
	if metadata == nil {
		return nil, errors.Join(errors.New("unable to allocate metadata dictionary"), writer.Close())
	}
	for _, entry := range options.metadata {
		// Ignore artwork fields
		if entry.Key == MetaArtwork || entry.Key == "" || entry.Value == nil {
			continue
		}
		// Set dictionary entry
		if err := ff.AVUtil_dict_set(metadata, entry.Key, fmt.Sprint(entry.Value), ff.AV_DICT_APPEND); err != nil {
			return nil, errors.Join(err, writer.Close())
		}
	}

	// Set metadata, write the header
	// Metadata ownership is transferred to the output context
	writer.output.SetMetadata(metadata)
	if err := ff.AVFormat_write_header(writer.output, nil); err != nil {
		return nil, errors.Join(err, writer.Close())
	} else {
		writer.header = true
	}

	// Return success
	return writer, nil
}

// Close a writer and release resources
func (w *Writer) Close() error {
	var result error

	// Write the trailer if the header was written
	if w.header {
		if err := ff.AVFormat_write_trailer(w.output); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Close encoders
	for _, encoder := range w.encoders {
		result = errors.Join(result, encoder.Close())
	}

	// Free output resources
	if w.output != nil {
		result = errors.Join(result, ff.AVFormat_close_writer(w.output))
	}

	// Free resources
	w.output = nil
	w.encoders = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a "stream" for encoding
func (w *Writer) Stream(stream int) *Encoder {
	for _, encoder := range w.encoders {
		if encoder.stream.Id() == stream {
			return encoder
		}
	}
	return nil
}

// Encode frames from all encoders, calling the callback function to encode
// the frame. If the callback function returns io.EOF then the encoding for
// that encoder is stopped after flushing. If the second callback is nil,
// then packets are written to the output.
func (w *Writer) Encode(in EncoderFrameFn, out EncoderPacketFn) error {
	if in == nil {
		return ErrBadParameter.With("nil in or out")
	}
	if out == nil {
		// By default, write packet to output
		out = func(pkt *ff.AVPacket, tb *ff.AVRational) error {
			return w.Write(pkt)
		}
	}

	// Initialise encoders
	encoders := make(map[int]*Encoder, len(w.encoders))
	for _, encoder := range w.encoders {
		stream := encoder.stream.Id()
		if _, exists := encoders[stream]; exists {
			return ErrBadParameter.Withf("duplicate stream %v", stream)
		}
		encoders[stream] = encoder

		// Initialize the encoder
		encoder.eof = false
	}

	// Continue until all encoders have returned io.EOF and have been flushed
	for {
		// No more encoding to do
		if len(encoders) == 0 {
			break
		}

		// TODO: We get the encoder with the lowest timestamp
		for stream, encoder := range encoders {
			var frame *ff.AVFrame
			var err error

			// Receive a frame if not EOF
			if !encoder.eof {
				frame, err = in(stream)
				if errors.Is(err, io.EOF) {
					encoder.eof = true
				} else if err != nil {
					return fmt.Errorf("stream %v: %w", stream, err)
				}
			}

			// Send a frame for encoding
			if err := encoder.Encode(frame, out); err != nil {
				return fmt.Errorf("stream %v: %w", stream, err)
			}

			// If eof then delete the encoder
			if encoder.eof {
				delete(encoders, stream)
			}
		}
	}

	// Return success
	return nil
}

// Write a packet to the output. If you intercept the packets in the
// Encode method, then you can use this method to write packets to the output.
func (w *Writer) Write(packet *ff.AVPacket) error {
	return ff.AVCodec_interleaved_write_frame(w.output, packet)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - Writer

func (w *writer_callback) Reader(buf []byte) int {
	if r, ok := w.w.(io.Reader); ok {
		if n, err := r.Read(buf); err != nil {
			return -1
		} else {
			return n
		}
	}
	return -1
}

func (w *writer_callback) Seeker(offset int64, whence int) int64 {
	whence = whence & ^ff.AVSEEK_FORCE
	seeker, ok := w.w.(io.ReadSeeker)
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

func (w *writer_callback) Writer(buf []byte) int {
	if n, err := w.w.Write(buf); err != nil {
		return -1
	} else {
		return n
	}
}
