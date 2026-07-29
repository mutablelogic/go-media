package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	httpclient "github.com/mutablelogic/go-media/gomedia/httpclient"
	profile "github.com/mutablelogic/go-media/profile/schema"
	taskencoder "github.com/mutablelogic/go-media/task/encoder"
	taskschema "github.com/mutablelogic/go-media/task/schema"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type EncoderCommands struct {
	Encode EncodeCmd `cmd:"" name:"encode" help:"Encode a media file into a new format." group:"ENCODER"`
}

type EncodeCmd struct {
	Path   string `arg:"" name:"path" type:"existingfile" help:"Path to the media file to encode."`
	Format string `name:"format" required:"" help:"Output container format, e.g. \"mp4\", \"mkv\", \"webm\"."`
	Stream bool   `name:"stream" help:"Stream progress events until the encode finishes, instead of returning immediately."`

	AudioCodec         string `name:"audio-codec" help:"Audio codec to encode with, e.g. \"aac\". If unset, no audio stream is produced."`
	AudioBitrate       uint64 `name:"audio-bitrate" help:"Audio bitrate in bits per second."`
	AudioSampleRate    uint64 `name:"audio-sample-rate" help:"Audio sample rate in Hz."`
	AudioSampleFormat  string `name:"audio-sample-format" help:"Audio sample format, e.g. \"fltp\"."`
	AudioChannelLayout string `name:"audio-channel-layout" help:"Audio channel layout, e.g. \"stereo\"."`

	VideoCodec       string  `name:"video-codec" help:"Video codec to encode with, e.g. \"libx264\". If unset, no video stream is produced."`
	VideoBitrate     uint64  `name:"video-bitrate" help:"Video bitrate in bits per second."`
	VideoWidth       uint64  `name:"video-width" help:"Frame width in pixels."`
	VideoHeight      uint64  `name:"video-height" help:"Frame height in pixels."`
	VideoPixelFormat string  `name:"video-pixel-format" help:"Video pixel format, e.g. \"yuv420p\"."`
	VideoFrameRate   float64 `name:"video-frame-rate" help:"Frames per second."`

	SubtitleCodec string `name:"subtitle-codec" help:"Subtitle codec to encode with, e.g. \"mov_text\". If unset, no subtitle stream is produced."`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *EncodeCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "Encode", func(ctx context.Context, client *httpclient.Client) error {
		f, err := os.Open(cmd.Path)
		if err != nil {
			return err
		}
		defer f.Close()

		req, err := cmd.request(f)
		if err != nil {
			return err
		}

		// Streaming prints each event as it arrives (the last of which
		// already carries the finished status), so there's nothing further
		// to print once Encode returns.
		var fn httpclient.EncodeEventFunc
		if cmd.Stream {
			fn = func(e *taskschema.Event) error {
				fmt.Println(types.Stringify(e))
				return nil
			}
		}

		status, err := client.Encode(ctx, req, nil, fn)
		if errors.Is(err, context.Canceled) {
			return nil
		} else if err != nil {
			return err
		}

		if !cmd.Stream {
			fmt.Println(types.Stringify(status))
		}
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// request builds the EncodeRequest cmd's flags describe - Output is always
// set; Audio/Video/Subtitle only when their respective *Codec flag was
// given, mirroring the task's own "each is optional, at least one required"
// contract.
func (cmd *EncodeCmd) request(f *os.File) (taskencoder.EncodeRequest, error) {
	req := taskencoder.EncodeRequest{Reader: f}

	output := profile.OutputWithName(cmd.Format)
	if output == nil {
		return req, gomedia.ErrBadParameter.Withf("output format %q is not available", cmd.Format)
	}
	req.Output = output

	if cmd.AudioCodec != "" {
		audio, err := profile.NewAudioProfile(cmd.AudioCodec)
		if err != nil {
			return req, err
		}
		if cmd.AudioBitrate > 0 {
			if err := audio.Set(profile.OptionBitrate, cmd.AudioBitrate); err != nil {
				return req, err
			}
		}
		if cmd.AudioSampleRate > 0 {
			if err := audio.Set(profile.OptionSampleRate, cmd.AudioSampleRate); err != nil {
				return req, err
			}
		}
		if cmd.AudioSampleFormat != "" {
			if err := audio.Set(profile.OptionSampleFormat, cmd.AudioSampleFormat); err != nil {
				return req, err
			}
		}
		if cmd.AudioChannelLayout != "" {
			if err := audio.Set(profile.OptionChannelLayout, cmd.AudioChannelLayout); err != nil {
				return req, err
			}
		}
		req.Audio = audio
	}

	if cmd.VideoCodec != "" {
		video, err := profile.NewVideoProfile(cmd.VideoCodec)
		if err != nil {
			return req, err
		}
		if cmd.VideoBitrate > 0 {
			if err := video.Set(profile.OptionBitrate, cmd.VideoBitrate); err != nil {
				return req, err
			}
		}
		if cmd.VideoWidth > 0 {
			if err := video.Set(profile.OptionWidth, cmd.VideoWidth); err != nil {
				return req, err
			}
		}
		if cmd.VideoHeight > 0 {
			if err := video.Set(profile.OptionHeight, cmd.VideoHeight); err != nil {
				return req, err
			}
		}
		if cmd.VideoPixelFormat != "" {
			if err := video.Set(profile.OptionPixelFormat, cmd.VideoPixelFormat); err != nil {
				return req, err
			}
		}
		if cmd.VideoFrameRate > 0 {
			if err := video.Set(profile.OptionFrameRate, cmd.VideoFrameRate); err != nil {
				return req, err
			}
		}
		req.Video = video
	}

	if cmd.SubtitleCodec != "" {
		subtitle, err := profile.NewSubtitleProfile(cmd.SubtitleCodec)
		if err != nil {
			return req, err
		}
		req.Subtitle = subtitle
	}

	return req, nil
}
