package audio

import (
	"fmt"
	"math"
	"runtime"
	"time"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type audioframe struct {
	sample_fmt     ffmpeg.AVSampleFormat
	rate           int
	channel_layout ffmpeg.AVChannelLayout
	nb_samples     int
	align          bool
	planar         bool
}

// Check interface compliance
var _ AudioFrame = (*audioframe)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioFrame(fmt AudioFormat, duration time.Duration, align bool) (*audioframe, error) {
	f := new(audioframe)

	if ffmpeg.AVUtil_av_sample_fmt_is_planar(f.sample_fmt) {
		f.align = align
		f.planar = true
	}

	// Set finalizer to panic if not closed
	runtime.SetFinalizer(f, audioframe_finalizer)

	// Set sample rate
	if fmt.Rate == 0 || fmt.Rate > math.MaxInt {
		return nil, ErrBadParameter.With("rate:", fmt.Rate)
	} else {
		f.rate = int(fmt.Rate)
	}

	// Set sample format
	if sample_fmt := toSampleFormat(fmt.Format); sample_fmt == ffmpeg.AV_SAMPLE_FMT_NONE || sample_fmt == ffmpeg.AV_SAMPLE_FMT_NB {
		return nil, ErrBadParameter.With("format:", fmt.Format)
	} else {
		f.sample_fmt = sample_fmt
	}

	// Set channel layout - default to mono
	if fmt.Layout == CHANNEL_LAYOUT_NONE {
		f.channel_layout = ffmpeg.AV_CHANNEL_LAYOUT_MONO
	} else {
		f.channel_layout = toChannelLayout(fmt.Layout)
	}

	// Set number of samples in a single channel
	if duration <= 0 {
		f.nb_samples = 0
	} else {
		f.nb_samples = int(duration * time.Duration(f.rate) / time.Second)
	}

	// Round up the number of samples

	// Allocate buffer
	//AVUtil_av_samples_alloc(buf, nil, buf.channels(), buf.nb_samples, buf.sample_fmt, align)
	// rate * duration / time.Second

	// Return success
	return f, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f *audioframe) String() string {
	str := "<AudioFrame"
	if f.align {
		str += " align"
	}
	if f.rate > 0 {
		str += fmt.Sprint(" rate=", f.rate)
		if f.sample_fmt != ffmpeg.AV_SAMPLE_FMT_NONE {
			str += fmt.Sprint(" sample_fmt=", f.sample_fmt)
		}
		if f.nb_samples > 0 {
			str += fmt.Sprint(" nb_samples=", f.nb_samples)
		}
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f *audioframe) Rate() int {
	return f.rate
}

func (f *audioframe) SampleFormat() SampleFormat {
	return fromSampleFormat(f.sample_fmt)
}

func (f *audioframe) ChannelLayout() ChannelLayout {
	return fromChannelLayout(f.channel_layout)
}

func (f *audioframe) Samples() int {
	return f.nb_samples
}

// Sample format
/*

	// Audio format
	AudioFormat() AudioFormat

	// Number of samples in a single channel
	Samples() int

	// Duration of the slide
	Duration() time.Duration

	// Number of audio channels
	Channels() int

	// Returns the samples for a specified channel, as array of bytes. For packed
	// audio format, the channel should be 0.
	Bytes(channel int) []byte
*/

func (f *audioframe) IsPlanar() bool {
	if f.sample_fmt != ffmpeg.AV_SAMPLE_FMT_NONE {
		return ffmpeg.AVUtil_av_sample_fmt_is_planar(f.sample_fmt)
	} else {
		return false
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func audioframe_finalizer(f *audioframe) {
	fmt.Println("swresample: audioframe_finalizer")
}
