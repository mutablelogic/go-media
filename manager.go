package media

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	// Package imports
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type manager struct {
}

type formatmeta struct {
	Name        string    `json:"name" writer:",width:25"`
	Description string    `json:"description" writer:",wrap,width:40"`
	Extensions  string    `json:"extensions,omitempty"`
	MimeTypes   string    `json:"mimetypes,omitempty" writer:",wrap,width:40"`
	MediaType   MediaType `json:"type,omitempty" writer:",wrap,width:21"`
}

type inputformat struct {
	formatmeta
	ctx *ff.AVInputFormat
}

type outputformat struct {
	formatmeta
	ctx *ff.AVOutputFormat
}

type device struct {
	Format      string    `json:"format"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Default     bool      `json:"default,omitempty"`
	MediaType   MediaType `json:"type,omitempty" writer:",wrap,width:21"`
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager() Manager {
	return new(manager)
}

func newInputFormat(ctx *ff.AVInputFormat, t MediaType) *inputformat {
	v := &inputformat{ctx: ctx}
	v.formatmeta.Name = strings.Join(v.Name(), " ")
	v.formatmeta.Description = v.Description()
	v.formatmeta.Extensions = strings.Join(v.Extensions(), " ")
	v.formatmeta.MimeTypes = strings.Join(v.MimeTypes(), " ")
	v.formatmeta.MediaType = INPUT | t
	return v
}

func newOutputFormat(ctx *ff.AVOutputFormat, t MediaType) *outputformat {
	v := &outputformat{ctx: ctx}
	v.formatmeta.Name = strings.Join(v.Name(), " ")
	v.formatmeta.Description = v.Description()
	v.formatmeta.Extensions = strings.Join(v.Extensions(), " ")
	v.formatmeta.MimeTypes = strings.Join(v.MimeTypes(), " ")
	v.formatmeta.MediaType = OUTPUT | t
	return v
}

func newInputDevice(ctx *ff.AVInputFormat, d *ff.AVDeviceInfo, t MediaType, def bool) *device {
	v := &device{}
	v.Format = ctx.Name()
	v.Name = d.Name()
	v.Description = d.Description()
	v.Default = def
	v.MediaType = INPUT | t
	return v
}

func newOutputDevice(ctx *ff.AVOutputFormat, d *ff.AVDeviceInfo, t MediaType, def bool) *device {
	v := &device{}
	v.Format = ctx.Name()
	v.Name = d.Name()
	v.Description = d.Description()
	v.Default = def
	v.MediaType = OUTPUT | t
	return v
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v *inputformat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ctx)
}

func (v *outputformat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ctx)
}

func (v *inputformat) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func (v *outputformat) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
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
	if t == NONE || t.Is(FILE) {
		var opaque uintptr
		for {
			demuxer := ff.AVFormat_demuxer_iterate(&opaque)
			if demuxer == nil {
				break
			}
			if matchesInput(demuxer, t, filter...) {
				result = append(result, newInputFormat(demuxer, FILE))
			}
		}
	}

	if t == NONE || t.Is(DEVICE) {
		// Iterate over all device inputs
		audio := ff.AVDevice_input_audio_device_first()
		for {
			if audio == nil {
				break
			}
			if matchesInput(audio, t, filter...) {
				result = append(result, newInputFormat(audio, AUDIO|DEVICE))
			}
			audio = ff.AVDevice_input_audio_device_next(audio)
		}

		video := ff.AVDevice_input_video_device_first()
		for {
			if video == nil {
				break
			}
			if matchesInput(video, t, filter...) {
				result = append(result, newInputFormat(video, VIDEO|DEVICE))
			}
			video = ff.AVDevice_input_video_device_next(video)
		}
	}

	// Return success
	return result
}

// Return the list of matching output formats, optionally filtering by name,
// extension or mimetype File extensions should be prefixed with a dot,
// e.g. ".mp4". The media type can be NONE (for any) or combinations of
// STREAM, DEVICE.
func (manager *manager) OutputFormats(t MediaType, filter ...string) []Format {
	var result []Format

	// Iterate over all output formats
	if t == NONE || t.Is(FILE) {
		var opaque uintptr
		for {
			muxer := ff.AVFormat_muxer_iterate(&opaque)
			if muxer == nil {
				break
			}
			if matchesOutput(muxer, t, filter...) {
				result = append(result, newOutputFormat(muxer, FILE))
			}
		}
	}

	// Iterate over all device outputs
	if t == NONE || t.Is(DEVICE) {
		audio := ff.AVDevice_output_audio_device_first()
		for {
			if audio == nil {
				break
			}
			if matchesOutput(audio, t, filter...) {
				result = append(result, newOutputFormat(audio, AUDIO|DEVICE))
			}
			audio = ff.AVDevice_output_audio_device_next(audio)
		}

		video := ff.AVDevice_output_video_device_first()
		for {
			if video == nil {
				break
			}
			if matchesOutput(video, t, filter...) {
				result = append(result, newOutputFormat(video, VIDEO|DEVICE))
			}
			video = ff.AVDevice_output_video_device_next(video)
		}
	}

	// Return success
	return result
}

// Return supported input devices for a given input format
func (manager *manager) InputDevices(format string) []Device {
	input := ff.AVFormat_find_input_format(format)
	if input == nil {
		return nil
	}

	device_list, err := ff.AVDevice_list_input_sources(input, format, nil)
	if err != nil {
		panic(err)
	}
	if device_list == nil {
		return nil
	}
	defer ff.AVDevice_free_list_devices(device_list)

	// Iterate over devices
	result := make([]Device, 0, device_list.NumDevices())
	for i, device := range device_list.Devices() {
		fmt.Println(i, device)
	}

	return result
}

// Return supported output devices for a given name
func (manager *manager) OutputDevices(format string) []Device {
	panic("TODO")
}

// Open a media file or device for reading, from a path or url.
func (manager *manager) Open(url string, format Format, opts ...string) (Media, error) {
	return Open(url, format, opts...)
}

// Open a media stream for reading.
func (manager *manager) Read(r io.Reader, format Format, opts ...string) (Media, error) {
	return NewReader(r, format, opts...)
}

// Create a media file for writing, from a path.
func (manager *manager) Create(string, Format) (Media, error) {
	return nil, ErrNotImplemented
}

// Create a media stream for writing.
func (manager *manager) Write(io.Writer, Format) (Media, error) {
	return nil, ErrNotImplemented
}

func (v *inputformat) Name() []string {
	return strings.Split(v.ctx.Name(), ",")
}

func (v *inputformat) Description() string {
	return v.ctx.LongName()
}

func (v *inputformat) Extensions() []string {
	result := []string{}
	for _, ext := range strings.Split(v.ctx.Extensions(), ",") {
		ext = strings.TrimSpace(ext)
		if ext != "" {
			result = append(result, "."+ext)
		}
	}
	return result
}

func (v *inputformat) MimeTypes() []string {
	result := []string{}
	for _, mimetype := range strings.Split(v.ctx.MimeTypes(), ",") {
		if mimetype != "" {
			result = append(result, mimetype)
		}
	}
	return result
}

func (v *inputformat) Type() MediaType {
	return INPUT
}

func (v *outputformat) Name() []string {
	return strings.Split(v.ctx.Name(), ",")
}

func (v *outputformat) Description() string {
	return v.ctx.LongName()
}

func (v *outputformat) Extensions() []string {
	result := []string{}
	for _, ext := range strings.Split(v.ctx.Extensions(), ",") {
		ext = strings.TrimSpace(ext)
		if ext != "" {
			result = append(result, "."+ext)
		}
	}
	return result
}

func (v *outputformat) MimeTypes() []string {
	result := []string{}
	for _, mimetype := range strings.Split(v.ctx.MimeTypes(), ",") {
		if mimetype != "" {
			result = append(result, mimetype)
		}
	}
	return result
}

func (v *outputformat) Type() MediaType {
	return OUTPUT
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func matchesInput(demuxer *ff.AVInputFormat, media_type MediaType, mimetype ...string) bool {
	// TODO: media_type

	// Match any
	if len(mimetype) == 0 && media_type == ANY {
		return true
	}
	// Match mimetype
	for _, mimetype := range mimetype {
		mimetype = strings.ToLower(strings.TrimSpace(mimetype))
		if slices.Contains(strings.Split(demuxer.Name(), ","), mimetype) {
			return true
		}
		if strings.HasPrefix(mimetype, ".") {
			ext := strings.TrimPrefix(mimetype, ".")
			if slices.Contains(strings.Split(demuxer.Extensions(), ","), ext) {
				return true
			}
		}
		if slices.Contains(strings.Split(demuxer.MimeTypes(), ","), mimetype) {
			return true
		}
	}
	// No match
	return false
}

func matchesOutput(muxer *ff.AVOutputFormat, media_type MediaType, filter ...string) bool {
	// TODO: media_type

	// Match any
	if len(filter) == 0 && media_type == ANY {
		return true
	}
	// Match mimetype
	for _, filter := range filter {
		if filter == "" {
			continue
		}
		filter = strings.ToLower(strings.TrimSpace(filter))
		if slices.Contains(strings.Split(muxer.Name(), ","), filter) {
			return true
		}
		if strings.HasPrefix(filter, ".") {
			if slices.Contains(strings.Split(muxer.Extensions(), ","), filter[1:]) {
				return true
			}
		}
		mt := strings.Split(muxer.MimeTypes(), ",")
		if slices.Contains(mt, filter) {
			return true
		}
	}
	// No match
	return false
}
