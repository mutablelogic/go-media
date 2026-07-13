package manager

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListAudioChannelLayouts returns standard FFmpeg channel layouts, optionally
// filtered by layout name and/or number of channels.
func (m *Media) ListAudioChannelLayouts(_ context.Context, req schema.ListAudioChannelLayoutRequest) (schema.ListAudioChannelLayoutResponse, error) {
	var iter uintptr
	response := make(schema.ListAudioChannelLayoutResponse, 0, 32)

	matches := func(ch *ff.AVChannelLayout) bool {
		if req.Name != "" {
			name, _ := ff.AVUtil_channel_layout_describe(ch)
			if name != req.Name {
				return false
			}
		}
		if req.NumChannels != 0 && ch.NumChannels() != req.NumChannels {
			return false
		}
		return true
	}

	for {
		ch := ff.AVUtil_channel_layout_standard(&iter)
		if ch == nil {
			break
		}
		if !matches(ch) {
			continue
		}
		if layout := schema.NewAudioChannelLayout(ch); layout != nil {
			response = append(response, *layout)
		}
	}

	return response, nil
}
