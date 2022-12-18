package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"syscall"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
	"golang.org/x/exp/slices"

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

// Open file for reading and return the media
func (m *manager) OpenFile(path string, format MediaFormat) (Media, error) {
	media, err := NewInputFile(path, format, func(media Media) error {
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

// Open URL for reading and return the media
func (m *manager) OpenURL(url string, format MediaFormat) (Media, error) {
	media, err := NewInputURL(url, format, func(media Media) error {
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

// Open media device with a specific name for reading and return it.
func (m *manager) OpenDevice(device string) (Media, error) {
	// Return device by name
	formats := m.MediaFormats(MEDIA_FLAG_DECODER|MEDIA_FLAG_DEVICE, device)
	if len(formats) == 0 {
		return nil, ErrNotFound.With(device)
	} else if len(formats) > 1 {
		return nil, ErrDuplicateEntry.With(device)
	}
	media, err := NewInputDevice(formats[0], func(media Media) error {
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

// Create an output device for writing and return it
func (m *manager) CreateDevice(device string) (Media, error) {
	// Return device by name
	formats := m.MediaFormats(MEDIA_FLAG_ENCODER|MEDIA_FLAG_DEVICE, device)
	if len(formats) == 0 {
		return nil, ErrNotFound.With(device)
	} else if len(formats) > 1 {
		return nil, ErrDuplicateEntry.With(device)
	}
	media, err := NewOutputDevice(formats[0], func(media Media) error {
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

// Create a new map for decoding
func (m *manager) Map(media Media, flags MediaFlag) (Map, error) {
	return NewMap(media, flags)
}

// Set the logging function for the manager
func (manager *manager) SetDebug(debug bool) {
	if debug {
		ffmpeg.AVUtil_av_log_set_level(ffmpeg.AV_LOG_DEBUG, manager.log)
	} else {
		ffmpeg.AVUtil_av_log_set_level(ffmpeg.AV_LOG_QUIET, manager.log)
	}
}

// Demux packets from a media file
func (manager *manager) Demux(ctx context.Context, media_map Map, fn DemuxFn) error {
	var result error

	// Get input
	input, ok := media_map.Input().(*input)
	if !ok || input == nil {
		return ErrBadParameter.With("input")
	}

	// Iterate over incoming packets, callback when packet should
	// be processed. Return if context is done
	packet := media_map.(*decodemap).Packet().(*packet)
	if packet == nil {
		return ErrBadParameter.With("packet")
	}
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
				// TODO: Flush calling avcoded_send_packet with nil
				break FOR_LOOP
			}
			if err := media_map.(*decodemap).Demux(ctx, packet, fn); err != nil {
				result = multierror.Append(result, err)
				break FOR_LOOP
			}
			packet.Release()
		}
	}

	// Close the map - cant be reused, so might make sense to create one
	// in this function?
	if err := media_map.(*decodemap).Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

// Decode packets into frames
func (manager *manager) Decode(ctx context.Context, media_map Map, p Packet, fn DecodeFn) error {
	stream := p.(*packet).StreamIndex()
	mapentry, exists := media_map.(*decodemap).context[stream]
	if !exists {
		return ErrBadParameter.With("stream")
	}
	decoder := mapentry.Decoder
	if decoder == nil || decoder.ctx == nil || decoder.frame == nil {
		return ErrBadParameter.With("decoder")
	}

	// Iterate through frames
	var result error
	for result == nil {
		// Receive frames from the packet
		err := ffmpeg.AVCodec_receive_frame(decoder.ctx, decoder.frame.ctx)
		if err == nil {
			// Resample
			if mapentry.Resampler != nil {
				if err := mapentry.Resampler.Resample(decoder.frame); err == nil {
					fmt.Println("Resample", mapentry.Resampler.Frame())
				} else {
					fmt.Println("Error", err)
				}
				// Release the frame
				mapentry.Resampler.Release()
			}
			err = fn(ctx, decoder.frame)
		}

		// TODO: Rescaler and Resampler

		// TODO: Encoder

		// TODO: Release frames for decoder, scaler, resampler, encoder for reuse
		decoder.frame.Release()

		// Check for errors
		if errors.Is(err, syscall.EAGAIN) {
			// Output is not available in this state - user must try to send new input
			break
		} else if err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

// Enumerate formats with MEDIA_FLAG_ENCODER, MEDIA_FLAG_DECODER, MEDIA_FLAG_FILE and
// MEDIA_FLAG_DEVICE flags. Use the filter argument to further filter by extension,
// name and mimetype
func (manager *manager) MediaFormats(flags MediaFlag, filter ...string) []MediaFormat {
	result := make([]MediaFormat, 0, 50) // Allocate an estimate of 50 media formats

	// Sanitize filter
	for i, name := range filter {
		filter[i] = strings.ToLower(strings.TrimSpace(name))
	}

	// If flags is MEDIA_FLAG_NONE then expand to all
	if flags == MEDIA_FLAG_NONE {
		flags = MEDIA_FLAG_ENCODER | MEDIA_FLAG_DECODER | MEDIA_FLAG_DEVICE | MEDIA_FLAG_FILE
	}
	// If flags does not contain MEDIA_FLAG_DEVICE OR MEDIA_FLAG_FILE, then expand to both
	if !(flags.Is(MEDIA_FLAG_DEVICE) || flags.Is(MEDIA_FLAG_FILE)) {
		flags |= MEDIA_FLAG_FILE | MEDIA_FLAG_DEVICE
	}

	// Append decoder input formats
	if flags.Is(MEDIA_FLAG_DECODER) {
		// File Formats
		if flags.Is(MEDIA_FLAG_FILE) {
			var opaque uintptr
			for {
				format := ffmpeg.AVFormat_av_demuxer_iterate(&opaque)
				if format == nil {
					break
				}
				result = appendMatchedFormat(result, filter, NewInputFormat(format, MEDIA_FLAG_DECODER|MEDIA_FLAG_FILE))
			}
		}
		// Devices
		if flags.Is(MEDIA_FLAG_DEVICE) {
			device := ffmpeg.AVDevice_av_input_audio_device_first()
			for device != nil {
				result = appendMatchedFormat(result, filter, NewInputFormat(device, MEDIA_FLAG_DECODER|MEDIA_FLAG_DEVICE|MEDIA_FLAG_AUDIO))
				device = device.AVDevice_av_input_audio_device_next()
			}
			device = ffmpeg.AVDevice_av_input_video_device_first()
			for device != nil {
				result = appendMatchedFormat(result, filter, NewInputFormat(device, MEDIA_FLAG_DECODER|MEDIA_FLAG_DEVICE|MEDIA_FLAG_VIDEO))
				device = device.AVDevice_av_input_video_device_next()
			}
		}
	}

	// Append encoder input formats
	if flags.Is(MEDIA_FLAG_ENCODER) {
		// File Formats
		if flags.Is(MEDIA_FLAG_FILE) {
			var opaque uintptr
			for {
				format := ffmpeg.AVFormat_av_muxer_iterate(&opaque)
				if format == nil {
					break
				}
				result = appendMatchedFormat(result, filter, NewOutputFormat(format, MEDIA_FLAG_ENCODER))
			}
		}
		// Devices
		if flags.Is(MEDIA_FLAG_DEVICE) {
			device := ffmpeg.AVDevice_av_output_audio_device_first()
			for device != nil {
				result = appendMatchedFormat(result, filter, NewOutputFormat(device, MEDIA_FLAG_ENCODER|MEDIA_FLAG_DEVICE|MEDIA_FLAG_AUDIO))
				device = device.AVDevice_av_output_audio_device_next()
			}
			device = ffmpeg.AVDevice_av_output_video_device_first()
			for device != nil {
				result = appendMatchedFormat(result, filter, NewOutputFormat(device, MEDIA_FLAG_ENCODER|MEDIA_FLAG_DEVICE|MEDIA_FLAG_VIDEO))
				device = device.AVDevice_av_output_video_device_next()
			}
		}
	}

	// Return formats
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (manager *manager) log(level ffmpeg.AVLogLevel, msg string, _ uintptr) {
	log.Println(level, msg)
}

func appendMatchedFormat(result []MediaFormat, filter []string, format MediaFormat) []MediaFormat {
	// If name is empty, append anyway
	if len(filter) == 0 || formatMatchesFilter(filter, format) {
		return append(result, format)
	} else {
		return result
	}
}

func formatMatchesFilter(filter []string, format MediaFormat) bool {
	for _, name := range filter {
		if strings.HasPrefix(name, ".") {
			return slices.Contains(format.Ext(), name)
		} else if slices.Contains(format.Name(), name) {
			return true
		} else if slices.Contains(format.MimeType(), name) {
			return true
		} else if slices.Contains(format.Ext(), "."+name) {
			return true
		}
	}
	// No match found
	return false
}
