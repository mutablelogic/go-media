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
