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

type ListAudioChannelLayoutRequest struct {
	Name        string `json:"name"`
	NumChannels int    `json:"num_channels"`
}

type ListAudioChannelLayoutResponse []AudioChannelLayout

type AudioChannelLayout struct {
	*ff.AVChannelLayout
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioChannelLayout(ch *ff.AVChannelLayout) *AudioChannelLayout {
	if ch == nil || !ff.AVUtil_channel_layout_check(ch) {
		return nil
	}
	return &AudioChannelLayout{AVChannelLayout: ch}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AudioChannelLayout) MarshalJSON() ([]byte, error) {
	if r.AVChannelLayout == nil {
		return json.Marshal(nil)
	}
	return r.AVChannelLayout.MarshalJSON()
}

func (r AudioChannelLayout) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r ListAudioChannelLayoutResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (AudioChannelLayout) Header() []string {
	return []string{"Name", "Channels", "Description"}
}

func (r AudioChannelLayout) Cell(col int) string {
	switch col {
	case 0:
		return r.Name()
	case 1:
		return strconv.Itoa(r.NumChannels())
	case 2:
		return r.Description()
	default:
		return ""
	}
}

func (AudioChannelLayout) Width(col int) int {
	switch col {
	case 0:
		return 24
	case 1:
		return 10
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r AudioChannelLayout) Name() string {
	if r.AVChannelLayout == nil {
		return ""
	}
	name, _ := ff.AVUtil_channel_layout_describe(r.AVChannelLayout)
	return name
}

func (r AudioChannelLayout) Description() string {
	if r.AVChannelLayout == nil {
		return ""
	}

	parts := make([]string, 0, r.NumChannels())
	for i := 0; i < r.NumChannels(); i++ {
		ch := ff.AVUtil_channel_layout_channel_from_index(r.AVChannelLayout, i)
		desc, err := ff.AVUtil_channel_description(ch)
		if err != nil || desc == "" {
			continue
		}
		parts = append(parts, desc)
	}

	return strings.Join(parts, ", ")
}
