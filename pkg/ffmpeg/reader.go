package ffmpeg

import (
	"errors"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type reader struct {
	input    *ff.AVFormatContext
	avio     *ff.AVIOContextEx
	decoders map[int]*decoder
	frame    *ff.AVFrame
}

type reader_callback struct {
	r io.Reader
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bufSize = 4096
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open a reader from a url
func OpenReader(url string, mimetype string) (*reader, error) {
	reader := new(reader)
	reader.decoders = make(map[int]*decoder)

	// TODO: mimetype input is currently ignored, format is always guessed

	// Open the stream
	if ctx, err := ff.AVFormat_open_url(url, nil, nil); err != nil {
		return nil, err
	} else {
		reader.input = ctx
	}

	// Find stream information and do rest of the initialization
	return reader.open()
}

// Create a new reader from an io.Reader
func NewReader(r io.Reader, mimetype string) (*reader, error) {
	reader := new(reader)
	reader.decoders = make(map[int]*decoder)

	// TODO: mimetype input is currently ignored, format is always guessed

	// Allocate the AVIO context
	reader.avio = ff.AVFormat_avio_alloc_context(bufSize, false, &reader_callback{r})
	if reader.avio == nil {
		return nil, errors.New("failed to allocate avio context")
	}

	// Open the stream
	if ctx, err := ff.AVFormat_open_reader(reader.avio, nil, nil); err != nil {
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

	// Create a frame for decoding
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		ff.AVFormat_free_context(r.input)
		ff.AVFormat_avio_context_free(r.avio)
		return nil, errors.New("failed to allocate frame")
	} else {
		r.frame = frame
	}

	// Return success
	return r, nil
}

// Close the reader
func (r *reader) Close() {
	// Free resources
	for _, decoder := range r.decoders {
		decoder.Close()
	}
	ff.AVUtil_frame_free(r.frame)
	ff.AVFormat_free_context(r.input)
	if r.avio != nil {
		ff.AVFormat_avio_context_free(r.avio)
	}

	// Release resources
	r.decoders = nil
	r.frame = nil
	r.input = nil
	r.avio = nil
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

type Decoder interface{}
type Packet interface{}
type Frame interface{}
type DecoderFunc func(Decoder, Packet) error
type FrameFunc func(Frame) error

// TODO: Frame should be a struct to access plane data and other properties
// TODO: Frame output may not include pts and time_base

// Demultiplex streams from the reader
func (r *reader) Demux(fn DecoderFunc) error {
	// Allocate a packet
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		return errors.New("failed to allocate packet")
	}
	defer ff.AVCodec_packet_free(packet)

	// Read packets
	for {
		if err := ff.AVFormat_read_frame(r.input, packet); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		stream := packet.StreamIndex()
		if decoder := r.decoders[stream]; decoder != nil {
			if err := fn(decoder, packet); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return err
			}
		}
		// Unreference the packet
		ff.AVCodec_packet_unref(packet)
	}

	// Flush the decoders
	for _, decoder := range r.decoders {
		if err := fn(decoder, nil); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

// Decode packets from the streams into frames
func (r *reader) Decode(fn FrameFunc) DecoderFunc {
	return func(codec Decoder, packet Packet) error {
		if packet != nil {
			// Submit the packet to the decoder
			if err := ff.AVCodec_send_packet(codec.(*decoder).codec, packet.(*ff.AVPacket)); err != nil {
				return err
			}
		} else {
			// Flush remaining frames
			if err := ff.AVCodec_send_packet(codec.(*decoder).codec, nil); err != nil {
				return err
			}
		}

		// get all the available frames from the decoder
		for {
			if err := ff.AVCodec_receive_frame(codec.(*decoder).codec, r.frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
				// Finished decoding packet or EOF
				return nil
			} else if err != nil {
				return err
			}

			// Resample or resize the frame, then pass back
			if frame, err := codec.(*decoder).re(r.frame); err != nil {
				return err
			} else if err := fn(frame); errors.Is(err, io.EOF) {
				// End early
				return nil
			} else if err != nil {
				return err
			}

			// Unreference the frame
			ff.AVUtil_frame_unref(r.frame)
		}
	}
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
