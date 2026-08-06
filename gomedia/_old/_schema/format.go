package schema

import (
	"encoding/json"
	"strconv"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListFormatRequest struct {
	Name     string `json:"name,omitempty" help:"Filter by format name (partial match)"`
	IsInput  *bool  `json:"is_input,omitempty" help:"Filter by input format (demuxer)"`
	IsOutput *bool  `json:"is_output,omitempty" help:"Filter by output format (muxer)"`
	IsDevice *bool  `json:"is_device,omitempty" help:"Filter by device format"`
}

type ListFormatResponse []Format

type Format struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	MimeTypes   []string `json:"mime_types,omitempty"`
	Extensions  []string `json:"extensions,omitempty"`
	IsInput     bool     `json:"is_input"`
	IsOutput    bool     `json:"is_output"`
	IsDevice    bool     `json:"is_device,omitempty"`
	Flags       []string `json:"flags,omitempty"`
	MediaTypes  []string `json:"media_types,omitempty"`

	DefaultVideoCodec    string `json:"default_video_codec,omitempty"`
	DefaultAudioCodec    string `json:"default_audio_codec,omitempty"`
	DefaultSubtitleCodec string `json:"default_subtitle_codec,omitempty"`

	Devices []Device       `json:"devices,omitempty"`
	Opts    []*ff.AVOption `json:"options,omitempty"`
}

type Device struct {
	Index       int      `json:"index"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsDefault   bool     `json:"is_default,omitempty"`
	MediaTypes  []string `json:"media_types,omitempty"`
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

	flags := input.Flags()
	var flagList []string
	if flagStr := flags.String(); flagStr != "" && flagStr != "AVFMT_NONE" {
		flagList = strings.Split(flagStr, "|")
	}

	var mimeTypes []string
	if mt := input.MimeTypes(); mt != "" {
		mimeTypes = strings.Split(mt, ",")
	}

	var extensions []string
	if ext := input.Extensions(); ext != "" {
		extensions = strings.Split(ext, ",")
	}

	var opts []*ff.AVOption
	if class := input.PrivClass(); class != nil {
		opts = ff.AVUtil_opt_list_from_class(class)
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
		Opts:        opts,
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

	flags := output.Flags()
	var flagList []string
	if flagStr := flags.String(); flagStr != "" && flagStr != "AVFMT_NONE" {
		flagList = strings.Split(flagStr, "|")
	}

	var mimeTypes []string
	if mt := output.MimeTypes(); mt != "" {
		mimeTypes = strings.Split(mt, ",")
	}

	var extensions []string
	if ext := output.Extensions(); ext != "" {
		extensions = strings.Split(ext, ",")
	}

	var opts []*ff.AVOption
	if class := output.PrivClass(); class != nil {
		opts = ff.AVUtil_opt_list_from_class(class)
	}

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
		Opts:                 opts,
	}
}

func NewDevice(info *ff.AVDeviceInfo, index int, isDefault bool) *Device {
	if info == nil {
		return nil
	}

	var mediaTypes []string
	for _, mt := range info.MediaTypes() {
		mediaTypes = append(mediaTypes, mediaTypeString(mt))
	}

	return &Device{
		Index:       index,
		Name:        info.Name(),
		Description: info.Description(),
		IsDefault:   isDefault,
		MediaTypes:  mediaTypes,
	}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (f *Format) SetDevices(devices []Device) {
	f.Devices = devices
}

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

func (r ListFormatResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (Format) Header() []string {
	return []string{"Name", "Kind", "Media", "Description"}
}

func (f Format) Cell(col int) string {
	switch col {
	case 0:
		return f.Name
	case 1:
		return f.Kind()
	case 2:
		return f.Media()
	case 3:
		return f.Description
	default:
		return ""
	}
}

func (Format) Width(col int) int {
	switch col {
	case 0:
		return 20
	case 1:
		return 12
	case 2:
		return 16
	default:
		return 0
	}
}

func (Device) Header() []string {
	return []string{"Index", "Name", "Media", "Default", "Description"}
}

func (d Device) Cell(col int) string {
	switch col {
	case 0:
		return strconv.Itoa(d.Index)
	case 1:
		return d.Name
	case 2:
		return d.Media()
	case 3:
		return strconv.FormatBool(d.IsDefault)
	case 4:
		return d.Description
	default:
		return ""
	}
}

func (Device) Width(col int) int {
	switch col {
	case 0:
		return 8
	case 1:
		return 24
	case 2:
		return 12
	case 3:
		return 8
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f Format) Kind() string {
	if f.IsInput && f.IsOutput {
		if f.IsDevice {
			return "io device"
		}
		return "input/output"
	}
	if f.IsInput {
		if f.IsDevice {
			return "input device"
		}
		return "input"
	}
	if f.IsOutput {
		if f.IsDevice {
			return "output device"
		}
		return "output"
	}
	if f.IsDevice {
		return "device"
	}
	return ""
}

func (f Format) Media() string {
	if len(f.MediaTypes) > 0 {
		return strings.Join(f.MediaTypes, ",")
	}

	media := make([]string, 0, 3)
	if f.DefaultVideoCodec != "" {
		media = append(media, "video")
	}
	if f.DefaultAudioCodec != "" {
		media = append(media, "audio")
	}
	if f.DefaultSubtitleCodec != "" {
		media = append(media, "subtitle")
	}

	return strings.Join(media, ",")
}

func (d Device) Media() string {
	return strings.Join(d.MediaTypes, ",")
}
