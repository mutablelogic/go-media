package sdl

import (
	"errors"
	"fmt"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	sdl "github.com/veandco/go-sdl2/sdl"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Audio struct {
	device sdl.AudioDeviceID
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	mapAudio = map[string]sdl.AudioFormat{
		"u8":   sdl.AUDIO_U8,
		"s8":   sdl.AUDIO_S8,
		"s16":  sdl.AUDIO_S16SYS,
		"flt":  sdl.AUDIO_F32SYS,
		"fltp": sdl.AUDIO_F32SYS,
	}
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (s *Context) NewAudio(par *ffmpeg.Par) (*Audio, error) {
	if !par.Type().Is(AUDIO) {
		return nil, errors.New("invalid audio parameters")
	}

	src_format := fmt.Sprint(par.SampleFormat())
	format, exists := mapAudio[src_format]
	if !exists {
		return nil, ErrBadParameter.Withf("unsupported sample format %q", src_format)
	}

	var desired, obtained sdl.AudioSpec
	desired.Freq = int32(par.Samplerate())
	desired.Format = format
	desired.Channels = uint8(par.ChannelLayout().NumChannels())
	desired.Samples = 1024
	//desired.Callback = s.AudioCallback

	if device, err := sdl.OpenAudioDevice("", false, &desired, &obtained, 0); err != nil {
		return nil, err
	} else {
		return &Audio{device}, nil
	}
}

func (a *Audio) Close() error {
	var result error

	// Close the audio device
	sdl.CloseAudioDevice(a.device)

	// Return any errors
	return result
}
