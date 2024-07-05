package ffmpeg

import (
	"encoding/json"
	"sort"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Codec ff.AVCodec

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newCodec(codec *ff.AVCodec) *Codec {
	return (*Codec)(codec)
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (codec *Codec) MarshalJSON() ([]byte, error) {
	return (*ff.AVCodec)(codec).MarshalJSON()
}

func (codec *Codec) String() string {
	data, _ := json.MarshalIndent((*ff.AVCodec)(codec), "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the type of codec
func (codec *Codec) Type() media.Type {
	switch (*ff.AVCodec)(codec).Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return media.AUDIO
	case ff.AVMEDIA_TYPE_VIDEO:
		return media.VIDEO
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return media.SUBTITLE
	}
	return media.NONE
}

// The name the codec is referred to by
func (codec *Codec) Name() string {
	return (*ff.AVCodec)(codec).Name()
}

// The description of the codec
func (codec *Codec) Description() string {
	return (*ff.AVCodec)(codec).LongName()
}

// Pixel formats supported by the codec. This is only valid for video codecs.
// The first pixel format is the default.
func (codec *Codec) PixelFormats() []string {
	pixfmts := (*ff.AVCodec)(codec).PixelFormats()
	result := make([]string, len(pixfmts))
	for i, pixfmt := range pixfmts {
		result[i] = ff.AVUtil_get_pix_fmt_name(pixfmt)
	}
	return result
}

// Sample formats supported by the codec. This is only valid for audio codecs.
// The first sample format is the default.
func (codec *Codec) SampleFormats() []string {
	samplefmts := (*ff.AVCodec)(codec).SampleFormats()
	result := make([]string, len(samplefmts))
	for i, samplefmt := range samplefmts {
		result[i] = ff.AVUtil_get_sample_fmt_name(samplefmt)
	}
	return result
}

// Sample rates supported by the codec. This is only valid for audio codecs.
// The first sample rate is the highest, sort the list in reverse order.
func (codec *Codec) SampleRates() []int {
	samplerates := (*ff.AVCodec)(codec).SupportedSamplerates()
	sort.Sort(sort.Reverse(sort.IntSlice(samplerates)))
	return samplerates
}

// Channel layouts supported by the codec. This is only valid for audio codecs.
func (codec *Codec) ChannelLayouts() []string {
	chlayouts := (*ff.AVCodec)(codec).ChannelLayouts()
	result := make([]string, 0, len(chlayouts))
	for _, chlayout := range chlayouts {
		name, err := ff.AVUtil_channel_layout_describe(&chlayout)
		if err != nil {
			continue
		}
		result = append(result, name)
	}
	return result
}

// Profiles supported by the codec. This is only valid for video codecs.
func (codec *Codec) Profiles() []string {
	profiles := (*ff.AVCodec)(codec).Profiles()
	result := make([]string, len(profiles))
	for i, profile := range profiles {
		result[i] = profile.Name()
	}
	return result
}
