package ffmpeg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
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

// EncoderFrameFn is a function which is called to receive a frame to encode. It should
// return nil to continue encoding or io.EOF to stop encoding.
type EncoderFrameFn func(int) (*Frame, error)

// EncoderPacketFn is a function which is called for each packet encoded, with
// the stream timebase.
type EncoderPacketFn func(*Packet) error

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

	// Try once more to get the output format
	var filename string
	if options.oformat == nil {
		if w_, ok := w.(*os.File); ok {
			filename = w_.Name()
			if err := OptOutputFormat(filename)(options); err != nil {
				return nil, err
			}
		}
	}

	// Bail out
	if options.oformat == nil {
		return nil, ErrBadParameter.Withf("invalid output format")
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
		if entry.Key() == MetaArtwork || entry.Key() == "" {
			continue
		}
		// Set dictionary entry
		if err := ff.AVUtil_dict_set(metadata, entry.Key(), entry.Value(), ff.AV_DICT_APPEND); err != nil {
			return nil, errors.Join(err, writer.Close())
		}
	}

	// Add artwork
	for _, entry := range options.metadata {
		// Ignore artwork fields
		if entry.Key() != MetaArtwork || len(entry.Bytes()) == 0 {
			continue
		}
		fmt.Println("TODO: Add artwork")
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

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Display the writer as a string
func (w *Writer) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.output)
}

// Display the writer as a string
func (w *Writer) String() string {
	data, _ := json.MarshalIndent(w, "", "  ")
	return string(data)
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
func (w *Writer) Encode(ctx context.Context, in EncoderFrameFn, out EncoderPacketFn) error {
	if in == nil {
		return ErrBadParameter.With("nil in or out")
	}
	if out == nil {
		// By default, write packet to output
		out = func(pkt *Packet) error {
			return w.Write(pkt)
		}
	}

	// Initialise encoders
	encoders := make(map[int]*Encoder, len(w.encoders))
	for _, encoder := range w.encoders {
		stream := encoder.stream.Index()
		if _, exists := encoders[stream]; exists {
			return ErrBadParameter.Withf("duplicate stream %v", stream)
		}
		encoders[stream] = encoder

		// Initialize the encoder
		encoder.eof = false
		encoder.next_pts = 0
	}

	// Continue until all encoders have returned io.EOF (or context cancelled)
	// and have been flushed
	for {
		// No more encoding to do
		if len(encoders) == 0 {
			break
		}

		// Mark as EOF if context is done
		select {
		case <-ctx.Done():
			// Mark all encoders as EOF to flush them
			for _, encoder := range encoders {
				encoder.eof = true
			}
			// Perform the encode
			if err := encode(in, out, encoders); err != nil {
				return err
			}
		default:
			// Perform the encode
			if err := encode(in, out, encoders); err != nil {
				return err
			}
		}
	}

	// Return error from context
	return ctx.Err()
}

// Write a packet to the output. If you intercept the packets in the
// Encode method, then you can use this method to write packets to the output.
func (w *Writer) Write(packet *Packet) error {
	return ff.AVCodec_interleaved_write_frame(w.output, (*ff.AVPacket)(packet))
}

// Returns -1 if a is before v
func compareNextPts(a, b *Encoder) int {
	return ff.AVUtil_compare_ts(a.next_pts, a.stream.TimeBase(), b.next_pts, b.stream.TimeBase())
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - Encoding

// Find encoder with the lowest timestamp, based on next_pts and timebase
// and send to the the EncoderFrameFn
func encode(in EncoderFrameFn, out EncoderPacketFn, encoders map[int]*Encoder) error {
	next_stream := -1
	var next_encoder *Encoder
	for stream, encoder := range encoders {
		if next_encoder == nil || compareNextPts(encoder, next_encoder) < 0 {
			next_encoder = encoder
			next_stream = stream
		}
	}

	// Receive a frame if not EOF
	var frame *Frame
	var err error
	if !next_encoder.eof {
		// Get the frame based on the id (rather than index) of the stream
		frame, err = in(next_encoder.stream.Id())
		if errors.Is(err, io.EOF) {
			next_encoder.eof = true
		} else if err != nil {
			return fmt.Errorf("stream %v: %w", next_encoder.stream.Id(), err)
		}
	}

	// Send a frame for encoding
	if err := next_encoder.Encode(frame, out); err != nil {
		return fmt.Errorf("stream %v: %w", next_encoder.stream.Id(), err)
	}

	// If eof then delete the encoder
	if next_encoder.eof {
		delete(encoders, next_stream)
		return nil
	}

	// Calculate the next PTS
	if frame != nil {
		next_encoder.next_pts = next_encoder.next_pts + next_encoder.nextPts(frame)
	}

	// Return success
	return nil
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
