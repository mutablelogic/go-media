package ffmpeg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"

	// Anonymous imports
	_ "image/jpeg" // Import for JPEG decoding
	_ "image/png"  // Import for PNG decoding

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Writer creates media files with multiple streams. Supports encoding frames or writing
// pre-encoded packets. For multiple streams, frames/packets are automatically interleaved
// by the muxer (AVFormat_interleaved_write_frame), but best results are achieved by
// sending frames in roughly temporal order to minimize buffering.
//
// # Multi-stream Encoding Best Practices
//
// When encoding multiple streams (e.g., video + audio), you have two options:
//
//  1. Temporal Interleaving (Recommended for streaming):
//     Compare timestamps across streams and encode the frame with the earliest timestamp.
//     This minimizes muxer buffering and is ideal for live streaming scenarios.
//     See Test_encode_multiple_streams_interleaved_mp4 for an example.
//
//  2. Batch by Stream (Simpler, works for files):
//     Encode all frames for one stream, then the next. The muxer will buffer and reorder.
//     This is simpler but uses more memory and adds latency.
//     See Test_encode_multiple_streams_mp4 for an example.
//
// Both approaches work correctly - the muxer handles interleaving automatically.
// Choose based on your use case (streaming vs file output, memory constraints, etc).
type Writer struct {
	output               *ff.AVFormatContext
	header               bool // Track if header was successfully written (for Close)
	encoders             []*encoder
	artworks             [][]byte   // Artworks to write after header (can be multiple)
	artworkStreamIndices []int      // Indices of artwork streams
	artworkOnce          sync.Once  // Ensure artwork is written only once (thread-safe)
	writeMutex           sync.Mutex // Protects concurrent writes to muxer
	copy                 bool       // Copy mode (remuxing without encoding)
}

func (w *Writer) writeInterleavedPacket(packet *Packet) error {
	w.writeMutex.Lock()
	// Use av_write_frame instead of av_interleaved_write_frame
	// because av_interleaved_write_frame buffers packets and we need
	// to understand the corruption issue first.
	err := ff.AVFormat_write_frame(w.output, (*ff.AVPacket)(packet))
	w.writeMutex.Unlock()
	return err
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
	if options.oformat == nil && url != "" {
		options.oformat = ff.AVFormat_guess_format("", url, "")
	}
	if options.oformat == nil {
		return nil, media.ErrBadParameter.With("unable to guess the output format")
	}

	// Allocate the output media context
	ctx, err := ff.AVFormat_create_file(url, options.oformat)
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
		return nil, media.ErrBadParameter.With("invalid output format")
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
	var result error

	// NOTE: options.streams is a map, so iteration order is nondeterministic.
	// Stream index assignment (0..N-1) depends on creation order; tests and callers
	// address streams by index, so we must create streams in a stable order.
	streamIDs := make([]int, 0, len(options.streams))
	for streamID := range options.streams {
		streamIDs = append(streamIDs, streamID)
	}
	sort.Ints(streamIDs)

	// Create encoders or copy streams based on copy flag
	if options.copy {
		// Copy mode: create streams without encoders (for remuxing)
		for _, stream := range streamIDs {
			par := options.streams[stream]
			// Create stream
			streamctx := ff.AVFormat_new_stream(writer.output, nil)
			if streamctx == nil {
				result = errors.Join(result, errors.New("failed to allocate stream"))
				continue
			}
			// Set stream ID
			streamctx.SetId(stream)
			// Copy codec parameters
			if err := ff.AVCodec_parameters_copy(streamctx.CodecPar(), &par.AVCodecParameters); err != nil {
				result = errors.Join(result, err)
				continue
			}
			// Set timebase if specified
			if par.timebase.Num() != 0 {
				streamctx.SetTimeBase(par.timebase)
			}
		}
	} else {
		// Encode mode: create codec contexts for each stream
		for _, stream := range streamIDs {
			par := options.streams[stream]
			encoder, err := newEncoder(writer.output, stream, par)
			if err != nil {
				result = errors.Join(result, err)
				continue
			}
			writer.encoders = append(writer.encoders, encoder)
		}
	}

	// Return any errors from stream/encoder creation
	if result != nil {
		return nil, errors.Join(result, writer.Close())
	}

	// Allocate metadata dictionary
	metadata := ff.AVUtil_dict_alloc()
	if metadata == nil {
		return nil, errors.Join(errors.New("unable to allocate metadata dictionary"), writer.Close())
	}
	// Note: No defer free - ownership transferred to output context via SetMetadata

	// Add metadata entries (but store artwork for later)
	for _, entry := range options.metadata {
		// Store artwork for later writing (after header)
		if entry.Key() == MetaArtwork {
			if data, ok := entry.meta.Value.([]byte); ok {
				writer.artworks = append(writer.artworks, data)
			}
			continue
		}
		// Ignore empty keys
		if entry.Key() == "" {
			continue
		}
		// Set dictionary entry
		if err := ff.AVUtil_dict_set(metadata, entry.Key(), entry.Value(), ff.AV_DICT_APPEND); err != nil {
			return nil, errors.Join(err, writer.Close())
		}
	}

	// Verify we have at least one stream
	if writer.output.NumStreams() == 0 {
		return nil, errors.Join(errors.New("no streams configured"), writer.Close())
	}

	// Create artwork streams for all artwork data provided (before header write)
	for _, artworkData := range writer.artworks {
		stream := ff.AVFormat_new_stream(writer.output, nil)
		if stream == nil {
			return nil, errors.Join(errors.New("failed to create artwork stream"), writer.Close())
		}

		// Detect image dimensions from artwork data
		img, _, err := image.DecodeConfig(bytes.NewReader(artworkData))
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to decode artwork image: %w", err), writer.Close())
		}

		// Detect codec ID from image content type
		codecID := codecIDFromImageData(artworkData)

		stream.CodecPar().SetCodecType(ff.AVMEDIA_TYPE_VIDEO)
		stream.CodecPar().SetCodecID(codecID)
		stream.CodecPar().SetWidth(img.Width)
		stream.CodecPar().SetHeight(img.Height)
		stream.SetDisposition(ff.AV_DISPOSITION_ATTACHED_PIC)
		writer.artworkStreamIndices = append(writer.artworkStreamIndices, int(stream.Index()))
	}

	// Set metadata, write the header
	// Metadata ownership is transferred to the output context
	writer.output.SetMetadata(metadata)
	if err := ff.AVFormat_write_header(writer.output, nil); err != nil {
		return nil, errors.Join(err, writer.Close())
	}
	writer.header = true

	// Return success
	return writer, nil
}

// writeArtwork writes all artwork packet data. This is called automatically by Write() and
// WritePackets() on the first packet write.
// Thread-safe: uses sync.Once to ensure artwork is written exactly once even with concurrent calls.
func (w *Writer) writeArtwork() error {
	var writeErr error
	w.artworkOnce.Do(func() {
		if len(w.artworks) == 0 || len(w.artworkStreamIndices) == 0 {
			return // No artwork to write
		}

		// Write each artwork as a separate packet
		for i, artworkData := range w.artworks {
			if i >= len(w.artworkStreamIndices) {
				break // Safety check
			}

			// Create packet for this artwork
			pkt := ff.AVCodec_packet_alloc()
			if pkt == nil {
				writeErr = errors.New("failed to allocate artwork packet")
				return
			}

			// Copy artwork data to packet
			if err := ff.AVCodec_packet_from_data(pkt, artworkData); err != nil {
				ff.AVCodec_packet_unref(pkt)
				ff.AVCodec_packet_free(pkt)
				writeErr = err
				return
			}
			pkt.SetStreamIndex(w.artworkStreamIndices[i])
			// Mark packet as key frame
			pkt.SetFlags(ff.AV_PKT_FLAG_KEY)

			// Write the artwork packet
			if err := ff.AVFormat_write_frame(w.output, pkt); err != nil {
				ff.AVCodec_packet_unref(pkt)
				ff.AVCodec_packet_free(pkt)
				writeErr = err
				return
			}

			// Release packet memory immediately after writing
			ff.AVCodec_packet_unref(pkt)
			ff.AVCodec_packet_free(pkt)
		}

		// Clear artwork data after writing
		w.artworks = nil
		w.artworkStreamIndices = nil
	})
	return writeErr
}

// Close a writer and release resources
func (w *Writer) Close() error {
	var result error

	// Write the trailer only if header was successfully written
	if w.header && w.output != nil {
		// Ensure no concurrent packet writes while flushing/trailer.
		w.writeMutex.Lock()
		// Flush any internally buffered/interleaving packets.
		_ = ff.AVFormat_interleaved_write_frame(w.output, nil)
		err := ff.AVFormat_write_trailer(w.output)
		w.writeMutex.Unlock()
		if err != nil {
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

	// Free resources and clear artwork data
	w.output = nil
	w.encoders = nil
	w.artworks = nil
	w.artworkStreamIndices = nil

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
func (w *Writer) Type() media.Type {
	return media.OUTPUT
}

// Return encoder for a given stream index
func (w *Writer) Stream(stream int) *encoder {
	for _, encoder := range w.encoders {
		if encoder.Stream() == stream {
			return encoder
		}
	}
	return nil
}

// Write a packet to the output (synchronous). The packet should have the stream index,
// PTS, DTS, duration, and timebase set correctly. Returns nil on success or error.
// Packets are automatically interleaved for proper ordering.
func (w *Writer) Write(packet *Packet) error {
	if packet == nil {
		return media.ErrBadParameter.With("nil packet")
	}
	if w.output == nil {
		return errors.New("writer is closed")
	}

	// Automatically write artwork on first packet if available
	if len(w.artworkStreamIndices) > 0 {
		if err := w.writeArtwork(); err != nil {
			return err
		}
	}

	err := w.writeInterleavedPacket(packet)
	if err != nil {
		return err
	}

	// Return success
	return nil
}

// WritePackets reads packets from a channel and writes them to the output (asynchronous).
// This method blocks until the channel is closed or an error occurs. Returns nil when
// the channel is closed normally, or an error if writing fails.
// Packets are automatically interleaved for proper ordering.
func (w *Writer) WritePackets(packets <-chan *Packet) error {
	if packets == nil {
		return media.ErrBadParameter.With("nil packet channel")
	}
	if w.output == nil {
		return errors.New("writer is closed")
	}

	// Automatically write artwork before first packet if available
	if len(w.artworkStreamIndices) > 0 {
		if err := w.writeArtwork(); err != nil {
			return err
		}
	}

	// Read packets from channel until closed
	for packet := range packets {
		if packet == nil {
			continue
		}
		err := w.writeInterleavedPacket(packet)
		if err != nil {
			return err
		}
		// Unref the packet after writing (caller should ref before sending)
		ff.AVCodec_packet_unref((*ff.AVPacket)(packet))
	}

	// Return success
	return nil
}

// EncodeFrame encodes a single frame and writes the resulting packet(s) to the output (synchronous).
// The frame should have the correct format for the stream's encoder. Pass nil to flush the encoder.
// Returns nil on success or error. Automatically writes artwork on first frame if available.
func (w *Writer) EncodeFrame(stream int, frame *Frame) error {
	if w.output == nil {
		return errors.New("writer is closed")
	}

	// Get encoder for this stream
	encoder := w.Stream(stream)
	if encoder == nil {
		return media.ErrBadParameter.With("no encoder for stream")
	}

	// Automatically write artwork on first frame if available
	if len(w.artworkStreamIndices) > 0 {
		if err := w.writeArtwork(); err != nil {
			return err
		}
	}

	// Encode frame and write packets
	return encoder.Encode(frame, func(pkt *Packet) error {
		if pkt == nil {
			// Flush signal, ignore
			return nil
		}
		return w.writeInterleavedPacket(pkt)
	})
}

// EncodeFrames encodes frames from a channel and writes the resulting packets to the output (asynchronous).
// This method blocks until the channel is closed or an error occurs. Returns nil when the channel is closed
// normally, or an error if encoding/writing fails. When the channel closes, the encoder is automatically
// flushed. Automatically writes artwork before first frame if available.
//
// Ownership: EncodeFrames assumes ownership of frames received from the channel and will Close()
// each frame after it has been sent to the encoder.
func (w *Writer) EncodeFrames(stream int, frames <-chan *Frame) error {
	if frames == nil {
		return media.ErrBadParameter.With("nil frame channel")
	}
	if w.output == nil {
		return errors.New("writer is closed")
	}

	// Get encoder for this stream
	encoder := w.Stream(stream)
	if encoder == nil {
		return media.ErrBadParameter.With("no encoder for stream")
	}

	// Automatically write artwork before first frame if available
	if len(w.artworkStreamIndices) > 0 {
		if err := w.writeArtwork(); err != nil {
			return err
		}
	}

	// Read frames from channel until closed
	for frame := range frames {
		if frame == nil {
			continue
		}

		// Work with a private copy to avoid aliasing frame data across goroutines
		copy, err := frame.Copy()
		if err != nil {
			return err
		}
		frame.Close()
		frame = copy

		err = encoder.Encode(frame, func(pkt *Packet) error {
			if pkt == nil {
				// Flush signal, ignore
				return nil
			}
			return w.writeInterleavedPacket(pkt)
		})
		frame.Close()

		if err != nil {
			return err
		}
	}

	// Flush encoder after channel closes
	return encoder.Encode(nil, func(pkt *Packet) error {
		if pkt == nil {
			// Flush signal, ignore
			return nil
		}
		return w.writeInterleavedPacket(pkt)
	})
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Detect codec ID from image data using content type detection
func codecIDFromImageData(data []byte) ff.AVCodecID {
	contentType := http.DetectContentType(data)

	switch contentType {
	case "image/jpeg":
		return ff.AV_CODEC_ID_MJPEG
	case "image/png":
		return ff.AV_CODEC_ID_PNG
	case "image/gif":
		return ff.AV_CODEC_ID_GIF
	case "image/bmp":
		return ff.AV_CODEC_ID_BMP
	case "image/webp":
		return ff.AV_CODEC_ID_WEBP
	default:
		// Default to JPEG for unknown image types
		return ff.AV_CODEC_ID_MJPEG
	}
}
