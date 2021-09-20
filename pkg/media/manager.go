package media

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	// Packages
	ffmpeg "github.com/djthorpe/go-media/sys/ffmpeg"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	// Debug will output debug messages on error channel
	Debug bool `yaml:"debug"`
}

type Manager struct {
	sync.Mutex
	in  []*MediaInput
	out []*MediaOutput
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	DefaultConfig = Config{Debug: false}
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManagerWithConfig(cfg Config, err chan<- error) (*Manager, error) {
	mgr := new(Manager)
	level := ffmpeg.AV_LOG_ERROR
	if cfg.Debug {
		level = ffmpeg.AV_LOG_DEBUG
	}
	ffmpeg.AVLogSetCallback(level, func(level ffmpeg.AVLogLevel, message string, userInfo uintptr) {
		select {
		case err <- NewMediaError(level, message):
			break
		default:
			break
		}
	})

	// Initialize format
	ffmpeg.AVFormatInit()

	// Return success
	return mgr, nil
}

func (mgr *Manager) Close() error {
	mgr.Mutex.Lock()
	defer mgr.Mutex.Unlock()

	// Close input streams
	var result error
	for _, in := range mgr.in {
		if err := in.Release(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Close output streams
	for _, out := range mgr.out {
		if err := out.Release(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Deinit
	ffmpeg.AVFormatDeinit()

	// Return to standard logging
	ffmpeg.AVLogSetCallback(0, nil)

	// Release resources
	mgr.in, mgr.out = nil, nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (mgr *Manager) String() string {
	str := "<manager"
	str += fmt.Sprintf(" version=%v", Version())
	if len(mgr.in) > 0 {
		str += fmt.Sprint(" in=", len(mgr.in))
	}
	if len(mgr.out) > 0 {
		str += fmt.Sprint(" out=", len(mgr.out))
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (mgr *Manager) OpenFile(path string) (*MediaInput, error) {
	// Clean up the path
	if abspath, err := filepath.Abs(path); err != nil {
		return nil, err
	} else {
		path = abspath
	}

	// Check to see if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotFound.With(path)
	} else if err != nil {
		return nil, err
	}

	// Create the media object and return it
	if ctx := ffmpeg.NewAVFormatContext(); ctx == nil {
		return nil, ErrInternalAppError.With("NewAVFormatContext")
	} else if err := ctx.OpenInput(path, nil); err != nil {
		return nil, err
	} else if in := NewMediaInput(ctx); in == nil {
		return nil, ErrInternalAppError.With("NewMediaInput")
	} else {
		mgr.Mutex.Lock()
		defer mgr.Mutex.Unlock()
		mgr.in = append(mgr.in, in)
		return in, nil
	}
}

func (mgr *Manager) OpenURL(url *url.URL) (*MediaInput, error) {
	// Check incoming parameters
	if url == nil {
		return nil, ErrBadParameter.With("OpenURL")
	}

	// Input
	if ctx := ffmpeg.NewAVFormatContext(); ctx == nil {
		return nil, ErrInternalAppError.With("NewAVFormatContext")
	} else if err := ctx.OpenInputUrl(url.String(), nil); err != nil {
		return nil, err
	} else if in := NewMediaInput(ctx); in == nil {
		return nil, ErrInternalAppError.With("NewMediaInput")
	} else {
		mgr.Mutex.Lock()
		defer mgr.Mutex.Unlock()
		mgr.in = append(mgr.in, in)
		return in, nil
	}
}

func (mgr *Manager) CreateFile(path string) (*MediaOutput, error) {
	// Clean up the path
	if abspath, err := filepath.Abs(path); err != nil {
		return nil, err
	} else {
		path = abspath
	}

	// Create file
	if ctx, err := ffmpeg.NewAVFormatOutputContext(path, nil); err != nil {
		return nil, err
	} else if out := NewMediaOutput(ctx); out == nil {
		return nil, ErrInternalAppError.With("NewMediaInput")
	} else {
		mgr.Mutex.Lock()
		defer mgr.Mutex.Unlock()
		mgr.out = append(mgr.out, out)
		return out, nil
	}
}

func (mgr *Manager) Release(f Media) error {
	if i, ok := f.(*MediaInput); ok {
		return mgr.ReleaseInput(i)
	} else if o, ok := f.(*MediaOutput); ok {
		return mgr.ReleaseOutput(o)
	} else {
		return ErrBadParameter.With("Release")
	}
}

func (mgr *Manager) ReleaseInput(f *MediaInput) error {
	// Remove from array
	for i, in := range mgr.in {
		if in == f {
			mgr.in = append(mgr.in[:i], mgr.in[i+1:]...)
			return f.Release()
		}
	}
	// Not found, return error
	return ErrInternalAppError.With("ReleaseInput")
}

func (mgr *Manager) ReleaseOutput(f *MediaOutput) error {
	// Remove from array
	for i, out := range mgr.out {
		if out == f {
			mgr.out = append(mgr.out[:i], mgr.out[i+1:]...)
			return f.Release()
		}
	}
	// Not found, return error
	return ErrInternalAppError.With("ReleaseOutput")
}

func (this *Manager) CodecByName(name string) *Codec {
	if name == "" {
		return nil
	}
	return NewCodec(ffmpeg.FindCodecByName(name))
}

func (this *Manager) Codecs(f ...MediaFlag) []*Codec {
	// Gather flags
	flags := MEDIA_FLAG_NONE
	for _, flag := range f {
		flags |= flag
	}

	// Enumerate all codecs
	result := make([]*Codec, 0, 100)
	for _, codec := range ffmpeg.AllCodecs() {
		result = append(result, NewCodec(codec))
	}
	if flags == MEDIA_FLAG_NONE {
		return result
	}

	// Filter by flags
	dst := 0
	for src, codec := range result {
		codecflags := codec.Flags()
		skip := false
		for _, test := range []MediaFlag{MEDIA_FLAG_VIDEO, MEDIA_FLAG_AUDIO, MEDIA_FLAG_SUBTITLE, MEDIA_FLAG_ENCODER, MEDIA_FLAG_DECODER} {
			if flags.Is(test) && !codecflags.Is(test) {
				skip = true
			}
		}
		if !skip {
			result[dst] = result[src]
			dst++
		}
	}

	// Return result
	return result[:dst]
}

func (mgr *Manager) Formats(f ...MediaFlag) []*Format {
	// Gather flags
	flags := MEDIA_FLAG_NONE
	for _, flag := range f {
		flags |= flag
	}

	// Enumerate all codecs
	result := make([]*Format, 0, 100)
	if flags == MEDIA_FLAG_NONE || flags&MEDIA_FLAG_ENCODER != 0 {
		for _, mux := range ffmpeg.AllMuxers() {
			result = append(result, NewOutputFormat(mux))
		}
	}
	if flags == MEDIA_FLAG_NONE || flags&MEDIA_FLAG_DECODER != 0 {
		for _, demux := range ffmpeg.AllDemuxers() {
			result = append(result, NewInputFormat(demux))
		}
	}

	// Filter by flags
	dst := 0
	for src, codec := range result {
		codecflags := codec.Flags()
		if flags.Is(MEDIA_FLAG_ENCODER) {
			if !codecflags.Is(MEDIA_FLAG_ENCODER) {
				continue
			}
		}
		if flags.Is(MEDIA_FLAG_DECODER) {
			if !codecflags.Is(MEDIA_FLAG_DECODER) {
				continue
			}
		}
		result[dst] = result[src]
		dst++
	}

	// Return result
	return result[:dst]
}
