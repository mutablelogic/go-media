package schema

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListAudioChannelLayoutRequest struct {
	Name        string `json:"name" kong:"help='Filter by channel layout name (e.g., stereo, 5.1)'"`
	NumChannels int    `json:"num_channels" kong:"help='Filter by number of channels'"`
}

type ListAudioChannelLayoutResponse []AudioChannelLayout

type AudioChannelLayout struct {
	*ff.AVChannelLayout `json:"-"`
	Name                string         `json:"name"`
	NumChannels         int            `json:"num_channels"`
	Order               string         `json:"order"`
	Channels            []AudioChannel `json:"channels"`
}

type AudioChannel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioChannelLayout(ch *ff.AVChannelLayout) *AudioChannelLayout {
	if ch == nil || !ff.AVUtil_channel_layout_check(ch) {
		return nil
	}

	// Get channel layout description
	description, err := ff.AVUtil_channel_layout_describe(ch)
	if err != nil {
		return nil
	}

	// Build channels array
	numChannels := ch.NumChannels()
	channels := make([]AudioChannel, numChannels)
	for i := 0; i < numChannels; i++ {
		avChannel := ff.AVUtil_channel_layout_channel_from_index(ch, i)
		name, _ := ff.AVUtil_channel_name(avChannel)
		desc, _ := ff.AVUtil_channel_description(avChannel)
		channels[i] = AudioChannel{
			Name:        name,
			Description: desc,
		}
	}

	// Return layout
	return &AudioChannelLayout{
		AVChannelLayout: ch,
		Name:            description,
		NumChannels:     numChannels,
		Order:           ch.Order().String(),
		Channels:        channels,
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AudioChannelLayout) String() string {
	data, err := json.MarshalIndent("TODO", "", "  ")
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
// PUBLIC METHODS

// QueryValues returns the URL query values for the request
func (r *ListAudioChannelLayoutRequest) QueryValues() url.Values {
	values := url.Values{}
	if name := strings.TrimSpace(r.Name); name != "" {
		values.Set("name", name)
	}
	if r.NumChannels > 0 {
		values.Set("num_channels", fmt.Sprint(r.NumChannels))
	}
	return values
}
