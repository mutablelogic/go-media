package schema

import (
	"net/url"
	"strconv"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type DeviceListRequest struct {
	Name     *string `json:"name,omitempty" help:"Filter by device name." placeholder:"Built-in Microphone" example:"Built-in Microphone"`
	Format   *string `json:"format,omitempty" help:"Filter by device format name." placeholder:"avfoundation" example:"avfoundation"`
	IsInput  *bool   `json:"is_input,omitempty" help:"Filter by input capability." example:"true"`
	IsOutput *bool   `json:"is_output,omitempty" help:"Filter by output capability." example:"false"`
}

type DeviceList []Device

type Device struct {
	Format      string   `json:"format" help:"Device format (driver) name." example:"avfoundation"`
	Index       int      `json:"index" help:"Device index within its format." example:"0"`
	Name        string   `json:"name" help:"Device name." example:"Built-in Microphone"`
	Description string   `json:"description,omitempty" help:"Device description." example:"Built-in Microphone"`
	IsDefault   bool     `json:"is_default,omitempty" help:"Whether this is the default device for its format." example:"true"`
	IsInput     bool     `json:"is_input" help:"Whether this device can be used as an input." example:"true"`
	IsOutput    bool     `json:"is_output" help:"Whether this device can be used as an output." example:"false"`
	MediaTypes  []string `json:"media_types,omitempty" help:"Media types supported by this device." example:"[\"audio\"]"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDevice(format string, isInput, isOutput bool, info *ff.AVDeviceInfo, index int, isDefault bool) *Device {
	if info == nil {
		return nil
	}

	mediaTypes := make([]string, 0, len(info.MediaTypes()))
	for _, mt := range info.MediaTypes() {
		mediaTypes = append(mediaTypes, CodecType(mt).String())
	}

	return &Device{
		Format:      format,
		Index:       index,
		Name:        info.Name(),
		Description: info.Description(),
		IsDefault:   isDefault,
		IsInput:     isInput,
		IsOutput:    isOutput,
		MediaTypes:  mediaTypes,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r DeviceListRequest) Query() url.Values {
	query := url.Values{}
	if r.Name != nil {
		query.Set("name", types.Value(r.Name))
	}
	if r.Format != nil {
		query.Set("format", types.Value(r.Format))
	}
	if r.IsInput != nil {
		query.Set("is_input", strconv.FormatBool(types.Value(r.IsInput)))
	}
	if r.IsOutput != nil {
		query.Set("is_output", strconv.FormatBool(types.Value(r.IsOutput)))
	}
	return query
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Device) String() string {
	return types.Stringify(r)
}

func (r DeviceList) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (Device) Header() []string {
	return []string{"Format", "Name", "Kind", "Media", "Default", "Description"}
}

func (r Device) Cell(col int) string {
	switch col {
	case 0:
		return r.Format
	case 1:
		return r.Name
	case 2:
		return r.Kind()
	case 3:
		return r.Media()
	case 4:
		return strconv.FormatBool(r.IsDefault)
	case 5:
		return r.Description
	default:
		return ""
	}
}

func (Device) Width(col int) int {
	switch col {
	case 0:
		return 14
	case 1:
		return 24
	case 2:
		return 12
	case 3:
		return 12
	case 4:
		return 8
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r Device) Kind() string {
	if r.IsInput && r.IsOutput {
		return "input/output"
	}
	if r.IsInput {
		return "input"
	}
	if r.IsOutput {
		return "output"
	}
	return ""
}

func (r Device) Media() string {
	return strings.Join(r.MediaTypes, ",")
}
