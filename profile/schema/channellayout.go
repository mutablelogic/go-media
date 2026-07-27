package schema

import (
	"net/url"
	"strconv"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ChannelLayoutListRequest struct {
	Name        *string `json:"name,omitempty" help:"Filter by channel layout name." placeholder:"5.1" example:"5.1"`
	NumChannels *uint64 `json:"num_channels,omitempty" help:"Filter by number of channels." placeholder:"6" example:"6"`
}

type ChannelLayoutList []ChannelLayout

type ChannelLayout struct {
	Name        string `json:"name" help:"Channel layout name." example:"5.1"`
	NumChannels int    `json:"num_channels" help:"Number of channels." example:"6"`

	// Private context for the channel layout
	ctx *ff.AVChannelLayout
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewChannelLayout(ch *ff.AVChannelLayout) *ChannelLayout {
	if ch == nil || !ff.AVUtil_channel_layout_check(ch) {
		return nil
	}
	name, _ := ff.AVUtil_channel_layout_describe(ch)
	return &ChannelLayout{
		Name:        name,
		NumChannels: ch.NumChannels(),
		ctx:         ch,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r ChannelLayoutListRequest) Query() url.Values {
	query := url.Values{}
	if r.Name != nil {
		query.Set("name", types.Value(r.Name))
	}
	if r.NumChannels != nil {
		query.Set("num_channels", strconv.FormatUint(types.Value(r.NumChannels), 10))
	}
	return query
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r ChannelLayout) String() string {
	return types.Stringify(r)
}

func (r ChannelLayoutList) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (ChannelLayout) Header() []string {
	return []string{"Name", "Channels"}
}

func (r ChannelLayout) Cell(col int) string {
	switch col {
	case 0:
		return r.Name
	case 1:
		return strconv.Itoa(r.NumChannels)
	default:
		return ""
	}
}

func (ChannelLayout) Width(col int) int {
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

// Context returns the underlying FFmpeg channel layout constant.
func (r ChannelLayout) Context() *ff.AVChannelLayout {
	return r.ctx
}
