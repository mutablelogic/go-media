package ffmpeg

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	ChannelLayout ff.AVChannelLayout
	Channel       ff.AVChannel
)

type jsonChannelLayout struct {
	Name        string     `json:"name"`
	NumChannels int        `json:"num_channels"`
	Order       string     `json:"order"`
	Channels    []*Channel `json:"channels"`
}

type jsonChannel struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newChannelLayout(channellayout *ff.AVChannelLayout) *ChannelLayout {
	if !ff.AVUtil_channel_layout_check(channellayout) {
		return nil
	}
	return (*ChannelLayout)(channellayout)
}

func newChannel(channel ff.AVChannel) *Channel {
	if channel == ff.AV_CHAN_NONE {
		return nil
	}
	return (*Channel)(&channel)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ch *ChannelLayout) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonChannelLayout{
		Name:        ch.Name(),
		NumChannels: ch.NumChannels(),
		Order:       ch.Order(),
		Channels:    ch.Channels(),
	})
}

func (ch *Channel) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonChannel{
		Name:        ch.Name(),
		Description: ch.Description(),
	})
}

func (ch *ChannelLayout) String() string {
	data, _ := json.MarshalIndent(ch, "", "  ")
	return string(data)
}

func (ch *Channel) String() string {
	data, _ := json.MarshalIndent(ch, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES - CHANNEL LAYOUT

func (ch *ChannelLayout) Name() string {
	if desc, err := ff.AVUtil_channel_layout_describe((*ff.AVChannelLayout)(ch)); err != nil {
		return ""
	} else {
		return desc
	}
}

func (ch *ChannelLayout) NumChannels() int {
	return ff.AVUtil_get_channel_layout_nb_channels((*ff.AVChannelLayout)(ch))
}

func (ch *ChannelLayout) Channels() []*Channel {
	var result []*Channel
	for i := 0; i < ch.NumChannels(); i++ {
		channel := ff.AVUtil_channel_layout_channel_from_index((*ff.AVChannelLayout)(ch), i)
		if channel != ff.AV_CHAN_NONE {
			result = append(result, newChannel(channel))
		}
	}
	return result
}

func (ch *ChannelLayout) Order() string {
	order := (*ff.AVChannelLayout)(ch).Order()
	switch order {
	case ff.AV_CHANNEL_ORDER_UNSPEC:
		return "unspecified"
	case ff.AV_CHANNEL_ORDER_NATIVE:
		return "native"
	case ff.AV_CHANNEL_ORDER_CUSTOM:
		return "custom"
	case ff.AV_CHANNEL_ORDER_AMBISONIC:
		return "ambisonic"
	}
	return order.String()
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES - CHANNEL

func (ch *Channel) Name() string {
	if desc, err := ff.AVUtil_channel_name((ff.AVChannel)(*ch)); err != nil {
		return "unknown"
	} else {
		return desc
	}
}

func (ch *Channel) Description() string {
	if desc, err := ff.AVUtil_channel_description((ff.AVChannel)(*ch)); err != nil {
		return ""
	} else {
		return desc
	}
}
