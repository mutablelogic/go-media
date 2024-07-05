package ffmpeg

import (
	"encoding/json"
	"fmt"
	"strings"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type metaFormat struct {
	Type media.Type `json:"type"`
	Name string     `json:"name"`
}

type Format struct {
	metaFormat
	Input   *ff.AVInputFormat  `json:"input,omitempty"`
	Output  *ff.AVOutputFormat `json:"output,omitempty"`
	Devices []*Device          `json:"devices,omitempty"`
}

type Device struct {
	metaDevice
}

type metaDevice struct {
	Name        string `json:"name" writer:",wrap,width:50"`
	Description string `json:"description" writer:",wrap,width:40"`
	Default     bool   `json:"default,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newInputFormats(demuxer *ff.AVInputFormat, t media.Type) []media.Format {
	names := strings.Split(demuxer.Name(), ",")
	result := make([]media.Format, 0, len(names))

	// Populate devices by name
	for _, name := range names {
		result = append(result, &Format{
			metaFormat: metaFormat{Type: t, Name: name},
			Input:      demuxer,
		})
	}

	// Get devices
	if t.Is(media.DEVICE) {
		list, err := ff.AVDevice_list_input_sources(demuxer, "", nil)
		if err == nil {
			fmt.Println(list)
		}
	}

	return result
}

func newOutputFormats(muxer *ff.AVOutputFormat, t media.Type) []media.Format {
	names := strings.Split(muxer.Name(), ",")
	result := make([]media.Format, 0, len(names))
	for _, name := range names {
		result = append(result, &Format{
			metaFormat: metaFormat{Type: t, Name: name},
			Output:     muxer,
		})
	}

	// Get devices
	if t.Is(media.DEVICE) {
		dict := ff.AVUtil_dict_alloc()
		defer ff.AVUtil_dict_free(dict)
		list, err := ff.AVDevice_list_output_sinks(muxer, "", dict)
		fmt.Println(err, list, dict)
	}

	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f *Format) String() string {
	data, _ := json.MarshalIndent(f, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f *Format) Type() media.Type {
	return f.metaFormat.Type
}

func (f *Format) Name() string {
	return f.metaFormat.Name
}
