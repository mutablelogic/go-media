package media

import (
	"io"
	"runtime"

	// Package imports
	version "github.com/mutablelogic/go-media/pkg/version"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type manager struct {
	opts
}

var _ Manager = (*manager)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(opt ...Opt) (Manager, error) {
	manager := new(manager)

	// Set default options
	manager.opts.level = ff.AV_LOG_WARNING

	// Apply other options
	for _, fn := range opt {
		if err := fn(&manager.opts); err != nil {
			return nil, err
		}
	}

	// Initialise the network
	ff.AVFormat_network_init()

	// Set logging
	ff.AVUtil_log_set_level(manager.level)
	if manager.callback != nil {
		ff.AVUtil_log_set_callback(func(level ff.AVLog, message string, userInfo any) {
			manager.callback(message)
		})
	}

	// Return success
	return manager, nil
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the list of matching input formats, optionally filtering by name,
// extension or mimetype File extensions should be prefixed with a dot,
// e.g. ".mp4". The media type can be NONE (for any) or combinations of
// STREAM, DEVICE.
func (manager *manager) InputFormats(t MediaType, filter ...string) []Format {
	var result []Format

	// Iterate over all input formats
	if t == ANY || t.Is(FILE) {
		var opaque uintptr
		for {
			demuxer := ff.AVFormat_demuxer_iterate(&opaque)
			if demuxer == nil {
				break
			}
			if matchesInput(demuxer, filter...) {
				result = append(result, newInputFormat(demuxer, FILE))
			}
		}
	}

	if t == ANY || t.Is(DEVICE) {
		// Iterate over all device inputs
		audio := ff.AVDevice_input_audio_device_first()
		for {
			if audio == nil {
				break
			}
			if matchesInput(audio, filter...) {
				result = append(result, newInputFormat(audio, AUDIO|DEVICE))
			}
			audio = ff.AVDevice_input_audio_device_next(audio)
		}

		video := ff.AVDevice_input_video_device_first()
		for {
			if video == nil {
				break
			}
			if matchesInput(video, filter...) {
				result = append(result, newInputFormat(video, VIDEO|DEVICE))
			}
			video = ff.AVDevice_input_video_device_next(video)
		}
	}

	// If no media filter, return all results
	if t == ANY {
		return result
	}

	result2 := make([]Format, 0, len(result))
	for _, format := range result {
		if format.Type().Is(t) {
			result2 = append(result2, format)
		}
	}

	// Return success
	return result2
}

// Return the list of matching output formats, optionally filtering by name,
// extension or mimetype File extensions should be prefixed with a dot,
// e.g. ".mp4". The media type can be NONE (for any) or combinations of
// STREAM, DEVICE.
func (manager *manager) OutputFormats(t MediaType, filter ...string) []Format {
	var result []Format

	// Iterate over all output formats
	if t == ANY || t.Is(FILE) {
		var opaque uintptr
		for {
			muxer := ff.AVFormat_muxer_iterate(&opaque)
			if muxer == nil {
				break
			}
			if matchesOutput(muxer, filter...) {
				result = append(result, newOutputFormat(muxer, FILE))
			}
		}
	}

	// Iterate over all device outputs
	if t == ANY || t.Is(DEVICE) {
		audio := ff.AVDevice_output_audio_device_first()
		for {
			if audio == nil {
				break
			}
			if matchesOutput(audio, filter...) {
				result = append(result, newOutputFormat(audio, AUDIO|DEVICE))
			}
			audio = ff.AVDevice_output_audio_device_next(audio)
		}

		video := ff.AVDevice_output_video_device_first()
		for {
			if video == nil {
				break
			}
			if matchesOutput(video, filter...) {
				result = append(result, newOutputFormat(video, VIDEO|DEVICE))
			}
			video = ff.AVDevice_output_video_device_next(video)
		}
	}

	// Return success
	return result
}

// Return supported devices for a given format. Sometimes
// (ie, AVFoundation) there is a option which provides the input
// devices and this function returns an empty string instead. Go figure!
func (manager *manager) Devices(format Format) []Device {
	var device_list *ff.AVDeviceInfoList
	if format == nil {
		return nil
	} else if !format.Type().Is(DEVICE) {
		return nil
	} else if format.Type().Is(INPUT) {
		if devices, err := ff.AVDevice_list_input_sources(format.(*inputformat).ctx, "", nil); err != nil {
			return nil
		} else {
			device_list = devices
		}
	} else if format.Type().Is(OUTPUT) {
		if devices, err := ff.AVDevice_list_output_sinks(format.(*outputformat).ctx, "", nil); err != nil {
			return nil
		} else {
			device_list = devices
		}
	}
	if device_list == nil {
		return []Device{}
	}
	defer ff.AVDevice_free_list_devices(device_list)

	// Iterate over devices
	result := make([]Device, 0, device_list.NumDevices())
	for _, device := range device_list.Devices() {
		// TODO: Default device
		result = append(result, newDevice(format.Name()[0], device, format.Type(), false))
	}
	return result
}

// Return all supported channel layouts
func (manager *manager) ChannelLayouts() []Metadata {
	var result []Metadata
	var iter uintptr
	for {
		ch := ff.AVUtil_channel_layout_standard(&iter)
		if ch == nil {
			break
		}
		if name, err := ff.AVUtil_channel_layout_describe(ch); err != nil {
			continue
		} else {
			result = append(result, newMetadata(name, ch))
		}
	}
	return result
}

// Return all supported sample formats
func (manager *manager) SampleFormats() []Metadata {
	var result []Metadata
	var opaque uintptr
	for {
		samplefmt := ff.AVUtil_next_sample_fmt(&opaque)
		if samplefmt == ff.AV_SAMPLE_FMT_NONE {
			break
		}
		if name := ff.AVUtil_get_sample_fmt_name(samplefmt); name != "" {
			result = append(result, newMetadata(name, samplefmt))
		}
	}
	return result
}

// Return all supported  pixel formats
func (manager *manager) PixelFormats() []Metadata {
	var result []Metadata
	var opaque uintptr
	for {
		pixfmt := ff.AVUtil_next_pixel_fmt(&opaque)
		if pixfmt == ff.AV_PIX_FMT_NONE {
			break
		}
		if name := ff.AVUtil_get_pix_fmt_name(pixfmt); name != "" {
			result = append(result, newMetadata(name, pixfmt))
		}
	}
	return result
}

// Return all supported codecs
func (manager *manager) Codecs() []Metadata {
	var result []Metadata
	var iter uintptr
	for {
		codec := ff.AVCodec_iterate(&iter)
		if codec == nil {
			break
		}
		if name := codec.Name(); name != "" {
			result = append(result, newMetadata(name, newCodec(codec)))
		}
	}
	return result
}

// Return audio parameters for encoding
// ChannelLayout, SampleFormat, Samplerate
func (manager *manager) AudioParameters(channels string, samplefmt string, samplerate int) (Parameters, error) {
	return newAudioParametersEx(channels, samplefmt, samplerate)
}

// Return video parameters for encoding
// Width, Height, PixelFormat, Framerate
func (manager *manager) VideoParameters(width int, height int, pixelfmt string) (Parameters, error) {
	return newVideoParametersEx(width, height, pixelfmt)
}

// Open a media file or device for reading, from a path or url.
func (manager *manager) Open(url string, format Format, opts ...string) (Media, error) {
	reader, err := newMedia(url, format, opts...)
	if err != nil {
		return nil, err
	}
	reader.force = manager.opts.force
	return reader, nil
}

// Open a media stream for reading.
func (manager *manager) Read(r io.Reader, format Format, opts ...string) (Media, error) {
	reader, err := newReader(r, format, opts...)
	if err != nil {
		return nil, err
	}
	reader.force = manager.opts.force
	return reader, err
}

// Create a media file for writing, from a url, path, or device.
func (manager *manager) Create(url string, format Format, metadata []Metadata, params ...Parameters) (Media, error) {
	return createMedia(url, format, metadata, params...)
}

// Create a media stream for writing.
func (manager *manager) Write(w io.Writer, format Format, metadata []Metadata, params ...Parameters) (Media, error) {
	return createWriter(w, format, metadata, params...)
}

// Return version information for the media manager as a set of metadata
func (manager *manager) Version() []Metadata {
	metadata := []Metadata{
		newMetadata("libavcodec_version", ff.AVCodec_version()),
		newMetadata("libavformat_version", ff.AVFormat_version()),
		newMetadata("libavutil_version", ff.AVUtil_version()),
		newMetadata("libavdevice_version", ff.AVDevice_version()),
		//		newMetadata("libavfilter_version", ff.AVFilter_version()),
		newMetadata("libswscale_version", ff.SWScale_version()),
		newMetadata("libswresample_version", ff.SWResample_version()),
	}
	if version.GitSource != "" {
		metadata = append(metadata, newMetadata("git_source", version.GitSource))
	}
	if version.GitBranch != "" {
		metadata = append(metadata, newMetadata("git_branch", version.GitBranch))
	}
	if version.GitTag != "" {
		metadata = append(metadata, newMetadata("git_tag", version.GitTag))
	}
	if version.GoBuildTime != "" {
		metadata = append(metadata, newMetadata("go_build_time", version.GoBuildTime))
	}
	if runtime.Version() != "" {
		metadata = append(metadata, newMetadata("go_version", runtime.Version()))
		metadata = append(metadata, newMetadata("go_arch", runtime.GOOS+"/"+runtime.GOARCH))
	}
	return metadata
}

// Log error messages
func (manager *manager) Errorf(f string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_ERROR, f, args...)
}

// Log warning messages
func (manager *manager) Warningf(f string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_WARNING, f, args...)
}

// Log info messages
func (manager *manager) Infof(f string, args ...any) {
	ff.AVUtil_log(nil, ff.AV_LOG_INFO, f, args...)
}
