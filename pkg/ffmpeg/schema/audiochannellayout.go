package schema

import (
	"encoding/json"

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
