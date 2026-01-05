package schema

import (
	"encoding/json"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListFormatRequest struct {
	Name     string `json:"name,omitempty"`      // Filter by format name (partial match)
	IsInput  *bool  `json:"is_input,omitempty"`  // Filter by input format (demuxer)
	IsOutput *bool  `json:"is_output,omitempty"` // Filter by output format (muxer)
	IsDevice *bool  `json:"is_device,omitempty"` // Filter by device format
}

type ListFormatResponse []Format

type Format struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	MimeTypes   []string `json:"mime_types,omitempty"`  // MIME types
	Extensions  []string `json:"extensions,omitempty"`  // File extensions
	IsInput     bool     `json:"is_input"`              // Is demuxer (input format)
	IsOutput    bool     `json:"is_output"`             // Is muxer (output format)
	IsDevice    bool     `json:"is_device,omitempty"`   // Is device format
	Flags       []string `json:"flags,omitempty"`       // Format flags
	MediaTypes  []string `json:"media_types,omitempty"` // Supported media types: "video", "audio", "subtitle"

	// Output format specific fields
	DefaultVideoCodec    string `json:"default_video_codec,omitempty"`
	DefaultAudioCodec    string `json:"default_audio_codec,omitempty"`
	DefaultSubtitleCodec string `json:"default_subtitle_codec,omitempty"`

	// Device specific fields
	Devices []Device `json:"devices,omitempty"` // Available devices (for device formats)
}

type Device struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsDefault   bool     `json:"is_default,omitempty"`
	MediaTypes  []string `json:"media_types,omitempty"` // "video", "audio", etc.
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewInputFormat(input *ff.AVInputFormat, isDevice bool) *Format {
	if input == nil {
		return nil
	}
	name := input.Name()
	if name == "" {
		return nil
	}

	// Parse flags
	flags := input.Flags()
	var flagList []string
	if flagStr := flags.String(); flagStr != "" && flagStr != "AVFMT_NONE" {
		flagList = strings.Split(flagStr, "|")
	}

	// Parse mime types
	var mimeTypes []string
	if mt := input.MimeTypes(); mt != "" {
		mimeTypes = strings.Split(mt, ",")
	}

	// Parse extensions
	var extensions []string
	if ext := input.Extensions(); ext != "" {
		extensions = strings.Split(ext, ",")
	}

	return &Format{
		Name:        name,
		Description: input.LongName(),
		MimeTypes:   mimeTypes,
		Extensions:  extensions,
		IsInput:     true,
		IsOutput:    false,
		IsDevice:    isDevice,
		Flags:       flagList,
	}
}

func NewOutputFormat(output *ff.AVOutputFormat, isDevice bool) *Format {
	if output == nil {
		return nil
	}
	name := output.Name()
	if name == "" {
		return nil
	}

	// Parse flags
	flags := output.Flags()
	var flagList []string
	if flagStr := flags.String(); flagStr != "" && flagStr != "AVFMT_NONE" {
		flagList = strings.Split(flagStr, "|")
	}

	// Parse mime types
	var mimeTypes []string
	if mt := output.MimeTypes(); mt != "" {
		mimeTypes = strings.Split(mt, ",")
	}

	// Parse extensions
	var extensions []string
	if ext := output.Extensions(); ext != "" {
		extensions = strings.Split(ext, ",")
	}

	// Get default codecs
	var videoCodec, audioCodec, subtitleCodec string
	if vc := output.VideoCodec(); vc != ff.AV_CODEC_ID_NONE {
		videoCodec = vc.Name()
	}
	if ac := output.AudioCodec(); ac != ff.AV_CODEC_ID_NONE {
		audioCodec = ac.Name()
	}
	if sc := output.SubtitleCodec(); sc != ff.AV_CODEC_ID_NONE {
		subtitleCodec = sc.Name()
	}

	// Derive media types from default codecs
	var mediaTypes []string
	if videoCodec != "" {
		mediaTypes = append(mediaTypes, "video")
	}
	if audioCodec != "" {
		mediaTypes = append(mediaTypes, "audio")
	}
	if subtitleCodec != "" {
		mediaTypes = append(mediaTypes, "subtitle")
	}

	return &Format{
		Name:                 name,
		Description:          output.LongName(),
		MimeTypes:            mimeTypes,
		Extensions:           extensions,
		IsInput:              false,
		IsOutput:             true,
		IsDevice:             isDevice,
		Flags:                flagList,
		MediaTypes:           mediaTypes,
		DefaultVideoCodec:    videoCodec,
		DefaultAudioCodec:    audioCodec,
		DefaultSubtitleCodec: subtitleCodec,
	}
}

func NewDevice(info *ff.AVDeviceInfo, isDefault bool) *Device {
	if info == nil {
		return nil
	}

	// Get media types
	var mediaTypes []string
	for _, mt := range info.MediaTypes() {
		mediaTypes = append(mediaTypes, mediaTypeString(mt))
	}

	return &Device{
		Name:        info.Name(),
		Description: info.Description(),
		IsDefault:   isDefault,
		MediaTypes:  mediaTypes,
	}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

// SetDevices adds device information to a format
func (f *Format) SetDevices(devices []Device) {
	f.Devices = devices
}

// AddMediaType adds a media type to the format if not already present
func (f *Format) AddMediaType(mediaType string) {
	for _, mt := range f.MediaTypes {
		if mt == mediaType {
			return
		}
	}
	f.MediaTypes = append(f.MediaTypes, mediaType)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Format) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r Device) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
