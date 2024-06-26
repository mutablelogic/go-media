package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type writer struct {
	t        MediaType
	output   *ff.AVFormatContext
	avio     *ff.AVIOContextEx
	metadata *ff.AVDictionary
	header   bool
	encoder  map[int]*encoder
}

type writer_callback struct {
	w io.Writer
}

var _ Media = (*writer)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create media from a url or device
func createMedia(url string, format Format, metadata []Metadata, params ...Parameters) (*writer, error) {
	writer := new(writer)
	writer.t = OUTPUT
	writer.encoder = make(map[int]*encoder, len(params))

	// If there are no streams, then return an error
	if len(params) == 0 {
		return nil, ErrBadParameter.With("no streams specified for encoder")
	}

	// Guess the output format
	var ofmt *ff.AVOutputFormat
	if format == nil && url != "" {
		ofmt = ff.AVFormat_guess_format("", url, "")
	} else if format != nil {
		ofmt = format.(*outputformat).ctx
	}
	if ofmt == nil {
		return nil, ErrBadParameter.With("unable to guess the output format")
	}

	// Allocate the output media context
	ctx, err := ff.AVFormat_create_file(url, ofmt)
	if err != nil {
		return nil, err
	} else {
		writer.output = ctx
	}

	// Add encoders and streams
	var result error
	for i, param := range params {
		// Stream Id from codec parameters, or use the index
		stream_id := param.Id()
		if stream_id <= 0 {
			stream_id = i + 1
		}
		encoder, err := newEncoder(ctx, stream_id, param)
		if err != nil {
			result = errors.Join(result, err)
		} else if _, exists := writer.encoder[stream_id]; exists {

		} else {
			writer.encoder[stream_id] = encoder
		}
	}

	// Return any errors from creating the streams
	if result != nil {
		return nil, errors.Join(result, writer.Close())
	}

	// Open the output file, if needed
	if !ctx.Flags().Is(ff.AVFMT_NOFILE) {
		w, err := ff.AVFormat_avio_open(url, ff.AVIO_FLAG_WRITE)
		if err != nil {
			return nil, errors.Join(err, writer.Close())
		} else {
			ctx.SetPb(w)
			writer.avio = w
		}
	}

	// Set metadata
	if len(metadata) > 0 {
		writer.metadata = ff.AVUtil_dict_alloc()
		if writer.metadata == nil {
			return nil, errors.Join(errors.New("unable to allocate metadata dictionary"), writer.Close())
		}
		for _, m := range metadata {
			// Ignore duration and artwork fields
			key := m.Key()
			if key == MetaArtwork || key == MetaDuration {
				continue
			}
			// Set dictionary entry
			if err := ff.AVUtil_dict_set(writer.metadata, key, fmt.Sprint(m.Value()), ff.AV_DICT_APPEND); err != nil {
				return nil, errors.Join(err, writer.Close())
			}
		}
		// TODO: Create artwork streams
	}

	// Write the header
	if err := ff.AVFormat_write_header(ctx, nil); err != nil {
		return nil, errors.Join(err, writer.Close())
	} else {
		writer.header = true
	}

	// Return success
	return writer, nil
}

// Create media from io.Writer
// TODO
func createWriter(w io.Writer, format Format, metadata []Metadata, params ...Parameters) (*writer, error) {
	return nil, ErrNotImplemented
}

func (w *writer) Close() error {
	var result error

	// Write the trailer if the header was written
	if w.header {
		if err := ff.AVFormat_write_trailer(w.output); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Close encoders
	for _, encoder := range w.encoder {
		result = errors.Join(result, encoder.Close())
	}

	// Free output resources
	if w.output != nil {
		// This calls avio_close(w.avio)
		result = errors.Join(result, ff.AVFormat_close_writer(w.output))
	}

	// Free resources
	if w.metadata != nil {
		ff.AVUtil_dict_free(w.metadata)
	}

	// Release resources
	w.encoder = nil
	w.metadata = nil
	w.avio = nil
	w.output = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Display the reader as a string
func (w *writer) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.output)
}

// Display the reader as a string
func (w *writer) String() string {
	data, _ := json.MarshalIndent(w, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (w *writer) Decoder(DecoderMapFunc) (Decoder, error) {
	return nil, ErrOutOfOrder.With("not an input stream")
}

func (w *writer) Mux(ctx context.Context, fn MuxFunc) error {
	// Check fn
	if fn == nil {
		return ErrBadParameter.With("nil mux function")
	}

	// Create a new map of encoders
	encoders := make(map[int]*encoder, len(w.encoder))
	for k, v := range w.encoder {
		encoders[k] = v
	}

FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		default:
			// Loop until no more encoders are available to send packets
			if len(encoders) == 0 {
				break FOR_LOOP
			}

			// Find the first encoder which should return a packet
			var next_encoder *encoder
			var next_stream int
			for stream, encoder := range encoders {
				// Initialise the next encoder
				if next_encoder == nil {
					next_encoder = encoder
					next_stream = stream
					continue
				}
				// Compare
				if !compareNextPts(next_encoder, encoder) {
					next_encoder = encoder
					next_stream = stream
				}
			}

			// Get a packet from the encoder
			packet, err := next_encoder.encode(fn)
			if errors.Is(err, io.EOF) {
				break FOR_LOOP
			} else if err != nil {
				return err
			} else if packet == nil {
				// Remove the encoder from the map
				delete(encoders, next_stream)
				continue FOR_LOOP
			}

			// Send the packet to the muxer
			//av_packet_rescale_ts(pkt, in_stream->time_base, out_stream->time_base);
			// Packet's stream_index field must be set to the index of the corresponding stream in s->streams.
			// The timestamps (pts, dts) must be set to correct values in the stream's timebase
			//  (unless the output format is flagged with the AVFMT_NOTIMESTAMPS flag, then they can be set
			// to AV_NOPTS_VALUE). The dts for subsequent packets in one stream must be strictly increasing
			// (unless the output format is flagged with the AVFMT_TS_NONSTRICT, then they merely have to
			// be nondecreasing). duration should also be set if known.
			if err := ff.AVCodec_interleaved_write_frame(w.output, packet); err != nil {
				return err
			}
		}
	}

	// Flush
	if err := ff.AVCodec_interleaved_write_frame(w.output, nil); err != nil {
		return err
	}

	// Return the context error, which will be nil if the loop ended normally
	return ctx.Err()
}

// Returns true if a.next_pts is greater than b.next_pts
func compareNextPts(a, b *encoder) bool {
	return ff.AVUtil_compare_ts(a.next_pts, a.stream.TimeBase(), b.next_pts, b.stream.TimeBase()) > 0
}

/*
		while (1) {
	        AVStream *in_stream, *out_stream;

	        ret = av_read_frame(ifmt_ctx, pkt);
	        if (ret < 0)
	            break;

	        in_stream  = ifmt_ctx->streams[pkt->stream_index];
	        if (pkt->stream_index >= stream_mapping_size ||
	            stream_mapping[pkt->stream_index] < 0) {
	            av_packet_unref(pkt);
	            continue;
	        }

	        pkt->stream_index = stream_mapping[pkt->stream_index];
	        out_stream = ofmt_ctx->streams[pkt->stream_index];
	        log_packet(ifmt_ctx, pkt, "in");

	        // copy packet
	        av_packet_rescale_ts(pkt, in_stream->time_base, out_stream->time_base);
	        pkt->pos = -1;
	        log_packet(ofmt_ctx, pkt, "out");

	        ret = av_interleaved_write_frame(ofmt_ctx, pkt);
	        // pkt is now blank (av_interleaved_write_frame() takes ownership of
	        // its contents and resets pkt), so that no unreferencing is necessary.
	        // This would be different if one used av_write_frame().
	        if (ret < 0) {
	            fprintf(stderr, "Error muxing packet\n");
	            break;
	        }
	    }
*/

// Return OUTPUT and combination of DEVICE and STREAM
func (w *writer) Type() MediaType {
	return OUTPUT
}

// Return the metadata for the media.
func (w *writer) Metadata(...string) []Metadata {
	// Not yet implemented
	return nil
}
