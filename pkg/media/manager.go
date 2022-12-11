package media

import (
	// Packages
	"context"
	"errors"
	"fmt"
	"io"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type manager struct {
	media map[Media]bool
}

// Ensure manager complies with Manager interface
var _ Manager = (*manager)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *manager {
	m := new(manager)
	m.media = make(map[Media]bool)
	return m
}

func (m *manager) Close() error {
	var result error

	// Close any opened media files
	var keys []Media
	for media := range m.media {
		keys = append(keys, media)
	}
	for _, media := range keys {
		if err := media.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Open media for reading and return it
func (m *manager) OpenFile(path string) (Media, error) {
	media, err := NewInputFile(path, func(media Media) error {
		delete(m.media, media)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Add to map
	m.media[media] = true

	// Return success
	return media, nil
}

// Create media for writing and return it
func (m *manager) CreateFile(path string) (Media, error) {
	media, err := NewOutputFile(path, func(media Media) error {
		delete(m.media, media)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Add to map
	m.media[media] = true

	// Return success
	return media, nil
}

// Set the logging function for the manager
func (manager *manager) SetDebug(debug bool) {
	if debug {
		ffmpeg.AVUtil_av_log_set_level(ffmpeg.AV_LOG_DEBUG, manager.log)
	} else {
		ffmpeg.AVUtil_av_log_set_level(ffmpeg.AV_LOG_QUIET, manager.log)
	}
}

// Decode packets from a media file
func (manager *manager) Decode(ctx context.Context, media Media) error {
	var result error

	// Ensure media is an input, not output
	input, ok := media.(*input)
	if !ok || input == nil || input.ctx == nil {
		return ErrBadParameter.With("media")
	}

	// Create a packet to contain the data
	packet := ffmpeg.AVCodec_av_packet_alloc()
	if packet == nil {
		return ErrInternalAppError
	}
	defer ffmpeg.AVCodec_av_packet_free(&packet)

	// Iterate over incoming packets, callback when packet should
	// be processed. Return if context is done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := ffmpeg.AVFormat_av_read_frame(input.ctx, packet); err != nil {
				if !errors.Is(err, io.EOF) {
					result = multierror.Append(result, err)
				}
				break FOR_LOOP
			}
			fmt.Println("PACKET", packet)
			ffmpeg.AVCodec_av_packet_unref(packet)
		}
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (manager *manager) log(level ffmpeg.AVLogLevel, msg string, _ uintptr) {
	fmt.Println("LOG=", level, msg)
}

/*
// Iterate over packets in the input stream
func (m *MediaInput) Read(ctx context.Context, streams []int, fn DecodeIteratorFunc) error {
	if fn == nil || m.ctx == nil {
		return ErrBadParameter.With("Read")
	}
	if len(streams) == 0 {
		for index := range m.s {
			streams = append(streams, index)
		}
	}

	// Create decode contexts
	var result error
	streammap := NewStreamMap()
	for _, i := range streams {
		if stream, exists := m.s[i]; !exists {
			result = multierror.Append(result, ErrNotFound.Withf("Stream with index %v", i))
		} else if err := streammap.Set(stream); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Bail out if any errors
	if result != nil {
		return result
	}

	// Create a packet
	packet := ffmpeg.NewAVPacket()
	if packet == nil {
		return ErrInternalAppError.With("NewAVPacket")
	}
	defer packet.Free()

	// Iterate over incoming packets, callback when packet should
	// be processed. Return if parent context is done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.ctx.ReadPacket(packet); err == io.EOF {
				// End of stream
				break FOR_LOOP
			} else if err != nil {
				return err
			} else if stream := streammap.Get(packet.Stream()); stream != nil {
				// Call decode function with packet
				err := fn(ctx, packet)
				packet.Release()
				if errors.Is(err, io.EOF) {
					// End of stream requested with no error
					break FOR_LOOP
				} else if err != nil {
					return err
				}
			}
		}
	}

	// Return success
	return nil
}
*/
