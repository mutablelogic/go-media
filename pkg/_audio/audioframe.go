package audio

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type audioframe struct {
	// The sample format of the audio frame
	sample_fmt ffmpeg.AVSampleFormat

	// The sample rate in Hz
	rate int

	// The channel layout
	layout ChannelLayout

	// Channel layout as ffmpeg.AVChannelLayout
	channel_layout *ffmpeg.AVChannelLayout

	// The number of channels
	channels []ffmpeg.AVChannel

	// Sample data. If planar, then each channel is a separate slice. If packed
	// then one slice is used.
	data []*byte

	// The numner of samples in a single channel
	nb_samples int

	// The number of samples in each data slice
	stride int

	// Whether alignment is required for planar data
	align bool

	// Whether the data is planar
	planar bool
}

// Check interface compliance
var _ AudioFrame = (*audioframe)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new audio frame with the expected format and duration
func NewAudioFrame(audio_fmt AudioFormat, duration time.Duration) (*audioframe, error) {
	return newAudioFrame(audio_fmt, duration, false, false)
}

// Create a new audio frame with the expected format and duration, as a planar
// frame (channels are separate). Set 'align' to true to align the data to boundaries
func NewAudioFramePlanar(audio_fmt AudioFormat, duration time.Duration, align bool) (*audioframe, error) {
	return newAudioFrame(audio_fmt, duration, true, align)
}

func newAudioFrame(audio_fmt AudioFormat, duration time.Duration, force_planar, align bool) (*audioframe, error) {
	f := new(audioframe)

	// Set finalizer to panic if not closed
	runtime.SetFinalizer(f, audioframe_finalizer)

	// Set sample rate
	if audio_fmt.Rate == 0 || audio_fmt.Rate > math.MaxInt {
		return nil, ErrBadParameter.With("rate:", audio_fmt.Rate)
	} else {
		f.rate = int(audio_fmt.Rate)
	}

	// Set sample format
	if sample_fmt := toSampleFormat(audio_fmt.Format); sample_fmt == ffmpeg.AV_SAMPLE_FMT_NONE || sample_fmt == ffmpeg.AV_SAMPLE_FMT_NB {
		return nil, ErrBadParameter.With("format:", audio_fmt.Format)
	} else {
		f.sample_fmt = sample_fmt
	}

	// Force to planar if requested
	if force_planar {
		f.sample_fmt = ffmpeg.AVUtil_av_get_planar_sample_fmt(f.sample_fmt)
	}

	// Set alignment and planar flags
	if ffmpeg.AVUtil_av_sample_fmt_is_planar(f.sample_fmt) {
		f.align = align
		f.planar = true
	} else {
		f.align = true
	}

	// Set channel layout - default to mono
	layout := ffmpeg.AV_CHANNEL_LAYOUT_MONO
	if audio_fmt.Layout == CHANNEL_LAYOUT_NONE {
		f.layout = CHANNEL_LAYOUT_MONO
		layout = ffmpeg.AV_CHANNEL_LAYOUT_MONO
	} else {
		f.layout = audio_fmt.Layout
		layout = toChannelLayout(audio_fmt.Layout)
	}
	f.channel_layout = &layout

	// Ensure valid channel layout
	if !ffmpeg.AVUtil_av_channel_layout_check(f.channel_layout) {
		return nil, ErrBadParameter.With("layout:", audio_fmt.Layout)
	}

	// Create array for audio channels and pointers to data
	if nb_channels := ffmpeg.AVUtil_av_get_channel_layout_nb_channels(f.channel_layout); nb_channels == 0 {
		return nil, ErrBadParameter.With("layout:", audio_fmt.Layout)
	} else {
		f.channels = make([]ffmpeg.AVChannel, nb_channels)

		// If planar, allocate array of pointers to data
		if f.planar {
			f.data = make([]*byte, nb_channels)
		} else {
			f.data = make([]*byte, 1)
		}
	}

	// Fill channels
	for i := range f.channels {
		if ch := ffmpeg.AVUtil_av_channel_layout_channel_from_index(f.channel_layout, i); ch == ffmpeg.AV_CHAN_NONE {
			return nil, ErrBadParameter.With("layout:", audio_fmt.Layout)
		} else {
			f.channels[i] = ch
		}
	}

	// Set number of samples in a single channel
	if duration <= 0 {
		f.nb_samples = 0
	} else {
		f.nb_samples = int(duration.Seconds() * float64(f.rate))
	}

	// Allocate buffer
	if f.nb_samples > 0 {
		if err := ffmpeg.AVUtil_av_samples_alloc(&f.data[0], &f.stride, len(f.channels), f.nb_samples, f.sample_fmt, toAlign(f.align)); err != nil {
			return nil, ErrInternalAppError.With("av_samples_alloc: ", err)
		}
	}

	// Return success
	return f, nil
}

func (f *audioframe) Close() error {
	var result error

	// Free any data
	ffmpeg.AVUtil_av_samples_free(&f.data[0])

	// Release resources
	f.rate = 0
	f.sample_fmt = ffmpeg.AV_SAMPLE_FMT_NONE
	f.nb_samples = 0
	f.channels = nil
	f.data = nil
	f.stride = 0

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f *audioframe) String() string {
	str := "<AudioFrame"
	if f.planar {
		str += " planar"
		if f.align {
			str += " align"
		}
	}
	if f.rate > 0 {
		str += fmt.Sprint(" rate=", f.rate)
		if f.sample_fmt != ffmpeg.AV_SAMPLE_FMT_NONE {
			str += fmt.Sprint(" sample_fmt=", f.sample_fmt)
		}
		if f.nb_samples > 0 {
			str += fmt.Sprint(" nb_samples=", f.nb_samples)
		}
		if len(f.channels) > 0 {
			str += fmt.Sprint(" channels=", f.channels)
		}
		if len(f.data) > 0 {
			str += fmt.Sprint(" data=", f.data)
		}
		if f.stride > 0 {
			str += fmt.Sprint(" stride=", f.stride)
		}
		str += fmt.Sprint(" duration=", f.Duration())
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the audio format of the frame
func (f *audioframe) AudioFormat() AudioFormat {
	return AudioFormat{
		Rate:   uint(f.rate),
		Format: fromSampleFormat(f.sample_fmt),
		Layout: f.layout,
	}
}

// Return true if the frame is planar, which means that the data is
// stored in separate data slices for each channel, as opposed to
// interleaved data in the zero-indexed slice.
func (f *audioframe) IsPlanar() bool {
	return f.planar
}

// Return the number of samples per channel in the frame
func (f *audioframe) Samples() int {
	return f.nb_samples
}

// Return an array of channel positions. For example, for a stereo
// frame, the array will contain [ CHANNEL_FRONT_LEFT, CHANNEL_FRONT_RIGHT ]
func (f *audioframe) Channels() []AudioChannel {
	result := make([]AudioChannel, len(f.channels))
	for i, ch := range f.channels {
		// TODO
		result[i] = AudioChannel(ch)
	}
	return result
}

// Return the duration of the audio frame, based on the number of
// samples and the sample rate
func (f *audioframe) Duration() time.Duration {
	return time.Second * time.Duration(f.nb_samples) / time.Duration(f.rate)
}

// Return the number of bytes per sample. For example, for a 16-bit
// sample format, this will return 2.
func (f *audioframe) BytesPerSample() int {
	return ffmpeg.AVUtil_av_get_bytes_per_sample(f.sample_fmt)
}

// Read samples from an io.Reader into a channel. Returns the number
// of bytes read and any error encountered. Returns io.EOF on end of
// the file
func (f *audioframe) Read(r io.Reader, ch int) (int, error) {
	data := f.Bytes(ch)
	if data == nil {
		return 0, ErrInternalAppError.With("invalid read for channel", ch)
	}
	return r.Read(data)
}

// Return slice of samples as bytes. ch should be zero unless planar audio.
func (f *audioframe) Bytes(ch int) []byte {
	var bytes []byte
	// Return nil if no data
	if ch < 0 || ch >= len(f.data) || f.data[ch] == nil {
		return nil
	}
	// Make a fake slice
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&bytes)))
	sliceHeader.Cap = ffmpeg.AVUtil_av_get_bytes_per_sample(f.sample_fmt) * f.nb_samples
	sliceHeader.Len = ffmpeg.AVUtil_av_get_bytes_per_sample(f.sample_fmt) * f.nb_samples
	sliceHeader.Data = uintptr(unsafe.Pointer(f.data[ch]))
	return bytes
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func audioframe_finalizer(f *audioframe) {
	if f.data != nil {
		panic("swresample: audioframe_finalizer: data not nil")
	}
}

func toAlign(align bool) int {
	if align {
		return 0
	} else {
		return 1
	}
}
