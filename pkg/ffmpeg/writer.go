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

	// Write the header
	if err := ff.AVFormat_write_header(writer.output, nil); err != nil {
		return nil, errors.Join(err, writer.Close())
	} else {
		writer.header = true
	}

	// Return success
	return writer, nil
}

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
		out = w.Write
	}

	// Initialise encoders
	encoders := make(map[int]*Encoder, len(w.encoders))
	for _, encoder := range w.encoders {
		stream := encoder.stream.Id()
		if _, exists := encoders[stream]; exists {
			return ErrBadParameter.Withf("duplicate stream %v", stream)
		}
		encoders[stream] = encoder
	}

	// Continue until all encoders have returned io.EOF
	for {
		// No more encoding to do
		if len(encoders) == 0 {
			break
		}
		for stream, encoder := range encoders {
			// Receive a frame for the encoder
			frame, err := in(stream)
			if errors.Is(err, io.EOF) {
				fmt.Println("EOF for frame on stream", stream)
				delete(encoders, stream)
			} else if err != nil {
				return fmt.Errorf("stream %v: %w", stream, err)
			} else if frame == nil {
				return fmt.Errorf("stream %v: nil frame received", stream)
			} else if err := encoder.Encode(frame, out); errors.Is(err, io.EOF) {
				fmt.Println("EOF for packet on stream", stream)
				delete(encoders, stream)
			} else if err != nil {
				return fmt.Errorf("stream %v: %w", stream, err)
			}
		}
	}

	// Return success
	return nil
}

// Write a packet to the output
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
