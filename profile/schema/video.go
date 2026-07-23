package schema

import (
	"encoding/json"
	"strings"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type VideoProfileMeta struct {
	Name        string          `json:"codec"  arg:"" required:""` // "libx264", "libx265", "copy", ...
	Bitrate     *uint64         `json:"bitrate,omitempty"`         // bps
	Profile     *string         `json:"profile,omitempty"`         // Codec profile; "high", "main", "baseline"
	Width       *uint64         `json:"width,omitempty"`           // Frame width in pixels
	Height      *uint64         `json:"height,omitempty"`          // Frame height in pixels
	PixelFormat *string         `json:"pixel_format,omitempty"`    // Video pixel format; "yuv420p", "nv12"
	FrameRate   *float64        `json:"frame_rate,omitempty"`      // Frames per second
	Opts        json.RawMessage `json:"options,omitempty"`         // Additional codec options

	// Unexported fields
	codec    *ff.AVCodec          `json:"-"` // Internal codec
	par      ff.AVCodecParameters `json:"-"` // Internal parameters
	timebase ff.AVRational        `json:"-"` // Internal timebase
	opts     map[string]Option    `json:"-"` // Internal codec options
}

type VideoProfile struct {
	Id uuid.UUID `json:"id,omitempty"` // Unique identifier for the video profile
	VideoProfileMeta
}

type VideoProfileUUID uuid.UUID

var _ Profile = (*VideoProfile)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	OptionWidth       = "width"
	OptionHeight      = "height"
	OptionPixelFormat = "pixel_format"
	OptionFrameRate   = "frame_rate"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewVideoProfile(codec string) (*VideoProfile, error) {
	// Create a new video profile with default values
	encoder := ff.AVCodec_find_encoder_by_name(codec)
	if encoder == nil {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not found", codec)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_VIDEO || encoder.IsEncoder() == false {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not a video encoding codec", codec)
	}

	self := &VideoProfile{
		VideoProfileMeta: VideoProfileMeta{
			Name:  encoder.Name(),
			codec: encoder,
			opts:  optionsForCodec(encoder),
		},
	}

	// Update internal codec parameters and timebase
	if err := self.setPar(); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r VideoProfile) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PROFILE INTERFACE

func (r VideoProfile) UUID() uuid.UUID {
	return r.Id
}

func (r VideoProfile) Type() CodecType {
	if r.codec == nil {
		return CodecType(ff.AVMEDIA_TYPE_UNKNOWN)
	}
	return CodecType(r.codec.Type())
}

func (r VideoProfile) Codec() *Codec {
	if r.codec == nil {
		return nil
	}
	return NewCodec(r.codec)
}

func (r VideoProfile) Par() *ff.AVCodecParameters {
	return types.Ptr(r.par)
}

func (r VideoProfile) TimeBase() *ff.AVRational {
	if r.timebase.Num() == 0 {
		return nil
	}
	return types.Ptr(r.timebase)
}

func (r VideoProfile) Options() json.RawMessage {
	return r.Opts
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - READER

// Expected column order: id, codec, bitrate, profile, width, height, pixel_format, frame_rate, opts.
func (r *VideoProfile) Scan(row pg.Row) error {
	if err := row.Scan(&r.Id, &r.Name, &r.Bitrate, &r.Profile, &r.Width, &r.Height, &r.PixelFormat, &r.FrameRate, &r.Opts); err != nil {
		return err
	}

	// Set context and options
	encoder := ff.AVCodec_find_encoder_by_name(r.Name)
	if encoder == nil {
		return gomedia.ErrBadParameter.Withf("codec %q is not found", r.Name)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_VIDEO || encoder.IsEncoder() == false {
		return gomedia.ErrBadParameter.Withf("codec %q is not a video encoding codec", r.Name)
	} else {
		r.codec = encoder
		r.opts = optionsForCodec(encoder)
	}

	// Set codec parameters and timebase
	if err := r.setPar(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - SELECTOR

func (r VideoProfileUUID) Select(bind *pg.Bind, op pg.Op) (string, error) {
	bind.Set("id", uuid.UUID(r))

	switch op {
	case pg.Get:
		return bind.Query("profile.video_get"), nil
	case pg.Delete:
		return bind.Query("profile.video_delete"), nil
	case pg.Update:
		return bind.Query("profile.video_update"), nil
	default:
		return "", gomedia.ErrInternalError.Withf("unsupported VideoProfileUUID operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - WRITER

// Insert binds values and returns the insert query for a video profile row.
func (r VideoProfileMeta) Insert(bind *pg.Bind) (string, error) {
	bind.Set("codec", r.Name)
	bind.Set(OptionBitrate, r.Bitrate)
	bind.Set(OptionProfile, r.Profile)
	bind.Set(OptionWidth, r.Width)
	bind.Set(OptionHeight, r.Height)
	bind.Set(OptionPixelFormat, r.PixelFormat)
	bind.Set(OptionFrameRate, r.FrameRate)
	if r.Opts == nil {
		bind.Set("opts", map[string]any{})
	} else {
		bind.Set("opts", r.Opts)
	}
	return bind.Query("profile.video_insert"), nil
}

// Update binds patch values for a video profile row update.
func (r VideoProfileMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")

	if bitrate := types.Value(r.Bitrate); bitrate > 0 {
		bind.Append("patch", `"`+OptionBitrate+`" = `+bind.Set(OptionBitrate, bitrate))
	}
	if value := strings.TrimSpace(types.Value(r.Profile)); value != "" {
		bind.Append("patch", `"`+OptionProfile+`" = `+bind.Set(OptionProfile, value))
	}
	if width := types.Value(r.Width); width > 0 {
		bind.Append("patch", `"`+OptionWidth+`" = `+bind.Set(OptionWidth, width))
	}
	if height := types.Value(r.Height); height > 0 {
		bind.Append("patch", `"`+OptionHeight+`" = `+bind.Set(OptionHeight, height))
	}
	if value := strings.TrimSpace(types.Value(r.PixelFormat)); value != "" {
		bind.Append("patch", `"`+OptionPixelFormat+`" = `+bind.Set(OptionPixelFormat, value))
	}
	if framerate := types.Value(r.FrameRate); framerate > 0 {
		bind.Append("patch", `"`+OptionFrameRate+`" = `+bind.Set(OptionFrameRate, framerate))
	}
	if r.Opts != nil {
		bind.Append("patch", `"opts" = `+bind.Set("opts", r.Opts))
	}
	if patch := bind.Join("patch", ", "); patch == "" {
		return gomedia.ErrBadParameter.With("no fields to update")
	} else {
		bind.Set("patch", patch)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - GET/SET OPTIONS

// Set a video profile option. If value is nil, the option is removed.
// TODO: If an error is returned, the option is not set and the profile is unchanged.
func (r *VideoProfileMeta) Set(name string, value any) error {
	// Check for existing option
	opt, exists := r.opts[name]
	if !exists {
		return gomedia.ErrBadParameter.Withf("option %q is not supported by codec %q", name, r.Name)
	}

	// Unmarshal the options JSON into a map
	var opts map[string]any
	if r.Opts == nil {
		opts = make(map[string]any)
	} else if err := json.Unmarshal(r.Opts, &opts); err != nil {
		return err
	}

	// Remove existing option
	if value == nil {
		switch name {
		case OptionBitrate:
			r.Bitrate = nil
		case OptionProfile:
			if len(r.codec.Profiles()) > 0 {
				r.Profile = nil
			} else {
				delete(opts, name)
			}
		case OptionWidth:
			r.Width = nil
		case OptionHeight:
			r.Height = nil
		case OptionPixelFormat:
			r.PixelFormat = nil
		case OptionFrameRate:
			r.FrameRate = nil
		default:
			delete(opts, name)
		}
	} else if value, err := opt.Validate(value); err != nil {
		return err
	} else {
		// Set the option value
		switch name {
		case OptionBitrate:
			r.Bitrate = types.Ptr(value.(uint64))
		case OptionProfile:
			// Some encoders (e.g. libx264, libx265) expose "profile" only as
			// their own private string option rather than the generic
			// AVCodecParameters.profile field, and don't declare anything
			// in codec.Profiles() — for those, defer to the codec's own
			// option dict instead of the dedicated numeric field.
			if len(r.codec.Profiles()) > 0 {
				r.Profile = types.Ptr(value.(string))
			} else {
				opts[name] = value
			}
		case OptionWidth:
			r.Width = types.Ptr(value.(uint64))
		case OptionHeight:
			r.Height = types.Ptr(value.(uint64))
		case OptionPixelFormat:
			r.PixelFormat = types.Ptr(value.(string))
		case OptionFrameRate:
			r.FrameRate = types.Ptr(value.(float64))
		default:
			opts[name] = value
		}
	}

	// Set the option value in the options map
	if optsJSON, err := json.Marshal(opts); err != nil {
		return err
	} else {
		r.Opts = optsJSON
	}

	// Set the internal codec parameters and timebase
	if err := r.setPar(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - GET/SET OPTIONS

func (r *VideoProfileMeta) setPar() error {
	// Check for codec
	if r.codec == nil {
		return gomedia.ErrInternalError.With("codec is not set")
	} else {
		r.par.SetCodecType(r.codec.Type())
		r.par.SetCodecID(r.codec.ID())
	}

	// Profile
	if id, err := resolveProfileID(r.codec, types.Value(r.Profile)); err != nil {
		return err
	} else {
		r.par.SetProfile(id)
	}

	// Bitrate
	if bitrate := types.Value(r.Bitrate); bitrate > 0 {
		r.par.SetBitRate(int64(bitrate))
	}

	// Pixel format
	if pixelformat := types.Value(r.PixelFormat); pixelformat != "" {
		if pixelformat_ := ff.AVUtil_get_pix_fmt(pixelformat); pixelformat_ == ff.AV_PIX_FMT_NONE {
			return gomedia.ErrBadParameter.Withf("unknown pixel format %q", pixelformat)
		} else {
			r.par.SetPixelFormat(pixelformat_)
		}
	}

	// Width, height
	if width := types.Value(r.Width); width > 0 {
		r.par.SetWidth(int(width))
	}
	if height := types.Value(r.Height); height > 0 {
		r.par.SetHeight(int(height))
	}

	// Frame rate: the timebase is the reciprocal of the frame rate, in the
	// same way an audio profile's timebase is the reciprocal of its sample
	// rate (one tick per frame instead of one tick per sample).
	if framerate := types.Value(r.FrameRate); framerate > 0 {
		r.timebase = ff.AVUtil_rational_invert(ff.AVUtil_rational_d2q(framerate, 0))
	}

	// Return success
	return nil
}
