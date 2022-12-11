package media

import (
	"context"
	"errors"
	"io"
	"log"

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
func (manager *manager) Decode(ctx context.Context, media Media, fn DecodeFn) error {
	var result error

	// Ensure media is an input, not output
	input, ok := media.(*input)
	if !ok || input == nil || input.ctx == nil {
		return ErrBadParameter.With("media")
	}

	// Create a packet to contain the data
	packet := NewPacket(func(i int) Stream {
		return input.streams[i]
	})
	if packet == nil {
		return ErrInternalAppError
	}
	defer packet.Close()

	// Iterate over incoming packets, callback when packet should
	// be processed. Return if context is done
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := ffmpeg.AVFormat_av_read_frame(input.ctx, packet.ctx); err != nil {
				if !errors.Is(err, io.EOF) {
					result = multierror.Append(result, err)
				}
				break FOR_LOOP
			}
			if err := fn(ctx, packet); err != nil {
				result = multierror.Append(result, err)
				break FOR_LOOP
			}
			packet.Release()
		}
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (manager *manager) log(level ffmpeg.AVLogLevel, msg string, _ uintptr) {
	log.Println(level, msg)
}
