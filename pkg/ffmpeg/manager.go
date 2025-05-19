package ffmpeg

import (
	"slices"
	"strings"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	opts []Opt
}

var _ media.Manager = (*Manager)(nil)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	manager *Manager
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new media manager which enumerates the available codecs, formats
// and devices
func NewManager(opt ...Opt) (*Manager, error) {
	var options opts

	// Return existing manager if it exists
	if manager == nil {
		manager = new(Manager)
	}

	// Set default options
	options.level = ff.AV_LOG_WARNING

	// Apply options
	for _, opt := range opt {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	// Set logging
	ff.AVUtil_log_set_level(options.level)
	if options.callback != nil {
		ff.AVUtil_log_set_callback(func(level ff.AVLog, message string, userInfo any) {
			options.callback(message)
		})
	}

	// Initialise network
	ff.AVFormat_network_init()

	// Set force flag - this is used to resample or resize decoded
	// frames even if the target format is the same as the source format
	if options.force {
		manager.opts = append(manager.opts, OptForce())
	}

	// Return success
	return manager, nil
}

// Open a media file or device for reading, from a path or url.
// If a format is specified, then the format will be used to open
// the file. You can add additional options to the open call as
// key=value pairs
/*
func (manager *Manager) Open(url string, format media.Format, opts ...string) (media.Media, error) {
	opt := append([]Opt{OptInputOpt(opts...)}, manager.opts...)
	if format != nil {
		opt = append(opt, OptInputFormat(format.Name()))
	}
	return Open(url, opt...)
}

// Open an io.Reader for reading. If a format is specified, then the
// format will be used to open the file. You can add additional options
// to the open call as key=value pairs
func (manager *Manager) NewReader(r io.Reader, format media.Format, opts ...string) (media.Media, error) {
	opt := append([]Opt{OptInputOpt(opts...)}, manager.opts...)
	if format != nil {
		opt = append(opt, OptInputFormat(format.Name()))
	}
	return NewReader(r, opt...)
}
*/

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - READER

func (manager *Manager) Open(url string, format media.Format, opts ...string) (media.Media, error) {
	o := append([]Opt{}, manager.opts[:]...)
	if format != nil {
		if format_, ok := format.(*Format); ok && format_.Input != nil {
			o = append(o, optInputFormat(format_))
		} else {
			return nil, media.ErrBadParameter.With("invalid input format")
		}
	}
	if len(opts) > 0 {
		o = append(o, OptInputOpt(opts...))
	}
	return Open(url, o...)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - FORMATS

func (manager *Manager) Formats(t media.Type, name ...string) []media.Format {
	// Create filters
	matchesInputFilter := func(demuxer *ff.AVInputFormat, filter string) bool {
		if strings.HasPrefix(filter, ".") {
			// By extension
			ext := strings.Split(demuxer.Extensions(), ",")
			for _, ext := range ext {
				if filter == "."+ext {
					return true
				}
			}
		} else if strings.Contains(filter, "/") {
			// By mimetype
			if slices.Contains(strings.Split(demuxer.MimeTypes(), ","), filter) {
				return true
			}
		} else {
			// By name
			if slices.Contains(strings.Split(demuxer.Name(), ","), filter) {
				return true
			}
		}
		return false
	}
	matchesOutputFilter := func(muxer *ff.AVOutputFormat, filter string) bool {
		if strings.HasPrefix(filter, ".") {
			// By extension
			ext := strings.Split(muxer.Extensions(), ",")
			for _, ext := range ext {
				if filter == "."+ext {
					return true
				}
			}
		} else if strings.Contains(filter, "/") {
			// By mimetype
			if slices.Contains(strings.Split(muxer.MimeTypes(), ","), filter) {
				return true
			}
		} else {
			// By name
			if slices.Contains(strings.Split(muxer.Name(), ","), filter) {
				return true
			}
		}
		return false
	}
	matchesInputFormat := func(demuxer *ff.AVInputFormat, t media.Type, filter ...string) bool {
		// Check for INPUT
		if !t.Is(media.INPUT) && !t.Is(media.ANY) {
			return false
		}
		if t.Is(media.DEVICE) {
			return false
		}
		if len(filter) == 0 {
			return true
		}
		for _, filter := range filter {
			if matchesInputFilter(demuxer, filter) {
				return true
			}
		}
		return false
	}
	matchesOutputFormat := func(muxer *ff.AVOutputFormat, t media.Type, filter ...string) bool {
		// Check for OUTPUT
		if !t.Is(media.OUTPUT) && !t.Is(media.ANY) {
			return false
		}
		if t.Is(media.DEVICE) {
			return false
		}
		if len(filter) == 0 {
			return true
		}
		for _, filter := range filter {
			if matchesOutputFilter(muxer, filter) {
				return true
			}
		}
		return false
	}
	matchesInputDevice := func(demuxer *ff.AVInputFormat, filter ...string) bool {
		if len(filter) == 0 {
			return true
		}
		for _, filter := range filter {
			if demuxer.Name() == filter {
				return true
			}
		}
		return false
	}
	matchesOutputDevice := func(muxer *ff.AVOutputFormat, filter ...string) bool {
		if len(filter) == 0 {
			return true
		}
		for _, filter := range filter {
			if muxer.Name() == filter {
				return true
			}
		}
		return false
	}

	// Iterate over all input formats
	var opaque uintptr
	result := []media.Format{}
	for {
		demuxer := ff.AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		if matchesInputFormat(demuxer, t, name...) {
			result = append(result, newInputFormats(demuxer, media.INPUT)...)
		}
	}

	// Iterate over all output formats
	var opaque2 uintptr
	for {
		muxer := ff.AVFormat_muxer_iterate(&opaque2)
		if muxer == nil {
			break
		}
		if matchesOutputFormat(muxer, t, name...) {
			result = append(result, newOutputFormats(muxer, media.OUTPUT)...)
		}
	}

	// Return if DEVICE is not requested
	if !t.Is(media.DEVICE) && !t.Is(media.ANY) {
		return result
	}

	// Iterate over all device inputs
	audio_input := ff.AVDevice_input_audio_device_first()
	for {
		if audio_input == nil {
			break
		}
		if matchesInputDevice(audio_input, name...) {
			result = append(result, newInputFormats(audio_input, media.INPUT|media.AUDIO|media.DEVICE)...)
		}
		audio_input = ff.AVDevice_input_audio_device_next(audio_input)
	}

	video_input := ff.AVDevice_input_video_device_first()
	for {
		if video_input == nil {
			break
		}
		if matchesInputDevice(video_input, name...) {
			result = append(result, newInputFormats(video_input, media.INPUT|media.VIDEO|media.DEVICE)...)
		}
		video_input = ff.AVDevice_input_video_device_next(video_input)
	}

	// Iterate over all device outputs
	audio_output := ff.AVDevice_output_audio_device_first()
	for {
		if audio_output == nil {
			break
		}
		if matchesOutputDevice(audio_output, name...) {
			result = append(result, newOutputFormats(audio_output, media.OUTPUT|media.AUDIO|media.DEVICE)...)
		}
		audio_output = ff.AVDevice_output_audio_device_next(audio_output)
	}

	video_output := ff.AVDevice_output_video_device_first()
	for {
		if video_output == nil {
			break
		}
		if matchesOutputDevice(video_output, name...) {
			result = append(result, newOutputFormats(video_output, media.OUTPUT|media.VIDEO|media.DEVICE)...)
		}
		video_output = ff.AVDevice_output_video_device_next(video_output)
	}

	// Return formats
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CODECS, PIXEL FORMATS, SAMPLE FORMATS AND CHANNEL
// LAYOUTS

// Return all supported sample formats
func (manager *Manager) SampleFormats() []media.Metadata {
	var result []media.Metadata
	var opaque uintptr
	for {
		samplefmt := ff.AVUtil_next_sample_fmt(&opaque)
		if samplefmt == ff.AV_SAMPLE_FMT_NONE {
			break
		}
		if sampleformat := newSampleFormat(samplefmt); sampleformat != nil {
			result = append(result, NewMetadata(sampleformat.Name(), sampleformat))
		}
	}
	return result
}

// Return all supported pixel formats
func (manager *Manager) PixelFormats() []media.Metadata {
	var result []media.Metadata
	var opaque uintptr
	for {
		pixfmt := ff.AVUtil_next_pixel_fmt(&opaque)
		if pixfmt == ff.AV_PIX_FMT_NONE {
			break
		}
		if pixelformat := newPixelFormat(pixfmt); pixelformat != nil {
			result = append(result, NewMetadata(pixelformat.Name(), pixelformat))
		}
	}
	return result
}

// Return standard channel layouts which can be used for audio
func (manager *Manager) ChannelLayouts() []media.Metadata {
	var result []media.Metadata
	var iter uintptr
	for {
		ch := ff.AVUtil_channel_layout_standard(&iter)
		if ch == nil {
			break
		}
		if channellayout := newChannelLayout(ch); channellayout != nil {
			result = append(result, NewMetadata(channellayout.Name(), channellayout))
		}
	}
	return result
}

// Return all supported codecs, of a specific type or all
// if ANY is used. If any names is provided, then only the codecs
// with those names are returned. Codecs can be AUDIO, VIDEO and
// SUBTITLE
func (manager *Manager) Codecs(t media.Type, name ...string) []media.Metadata {
	var iter uintptr

	// Filter to match codecs
	codecMatchesFilter := func(codec *Codec, t media.Type, names ...string) bool {
		if codec == nil {
			return false
		}
		if !(t == media.ANY || codec.Type().Is(t)) {
			return false
		}
		if len(name) > 0 && !slices.Contains(names, codec.Name()) {
			return false
		}
		return true
	}

	// Iterate over codecs
	result := []media.Metadata{}
	for {
		codec := ff.AVCodec_iterate(&iter)
		if codec == nil {
			break
		}
		codec_ := newCodec(codec)
		if codecMatchesFilter(codec_, t, name...) {
			result = append(result, NewMetadata(codec_.Name(), codec_))
		}
	}

	// Return matched codecs
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - LOGGING

// Log error messages
func (manager *Manager) Errorf(v string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_ERROR, v, args...)
}

// Log warning messages
func (manager *Manager) Warningf(v string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_WARNING, v)
}

// Log info messages
func (manager *Manager) Infof(v string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_INFO, v, args...)
}
