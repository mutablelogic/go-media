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
	Name        string         `json:"name"`
	NumChannels int            `json:"num_channels"`
	Order       string         `json:"order"`
	Channels    []AudioChannel `json:"channels"`
}

type AudioChannel struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioChannelLayout(ch *ff.AVChannelLayout) *AudioChannelLayout {
	if ch == nil || !ff.AVUtil_channel_layout_check(ch) {
		return nil
	}
	name, _ := ff.AVUtil_channel_layout_describe(ch)
	numChannels := ch.NumChannels()
	layout := &AudioChannelLayout{
		Name:        name,
		NumChannels: numChannels,
		Order:       ch.Order().String(),
	}
	channels := make([]AudioChannel, 0, numChannels)
	for i := 0; i < numChannels; i++ {
		channel := ff.AVUtil_channel_layout_channel_from_index(ch, i)
		channelName, err := ff.AVUtil_channel_name(channel)
		if err != nil {
			continue
		}
		channelDesc, _ := ff.AVUtil_channel_description(channel)
		channels = append(channels, AudioChannel{
			Index:       i,
			Name:        channelName,
			Description: channelDesc,
		})
	}
	layout.Channels = channels
	return layout
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AudioChannelLayout) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r AudioChannel) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
