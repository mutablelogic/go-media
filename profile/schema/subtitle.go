package schema

import (
	"encoding/json"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// SubtitleProfileMeta is deliberately minimal compared to Audio/VideoProfileMeta:
// subtitle codecs have no universal structural parameter analogous to audio's
// sample rate or video's width/height/frame rate — every sampled encoder
// (srt, ass, webvtt, mov_text, dvbsub, dvdsub) declares neither a static
// profile list nor any shared option. Codec-specific knobs (mov_text's
// "height", dvbsub's "min_bpp", dvdsub's "palette") flow through Opts like
// any other private codec option, the same as x264's crf/preset.
type SubtitleProfileMeta struct {
	Name string          `json:"codec"  arg:"" required:""` // "srt", "ass", "webvtt", "mov_text", "copy", ...
	Opts json.RawMessage `json:"options,omitempty"`         // Additional codec options

	// Unexported fields
	codec *ff.AVCodec          `json:"-"` // Internal codec
	par   ff.AVCodecParameters `json:"-"` // Internal parameters
	opts  map[string]Option    `json:"-"` // Internal codec options
}

type SubtitleProfile struct {
	Id uuid.UUID `json:"id,omitempty"` // Unique identifier for the subtitle profile
	SubtitleProfileMeta
}

type SubtitleProfileUUID uuid.UUID

var _ Profile = (*SubtitleProfile)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSubtitleProfile(codec string) (*SubtitleProfile, error) {
	// Create a new subtitle profile with default values
	encoder := ff.AVCodec_find_encoder_by_name(codec)
	if encoder == nil {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not found", codec)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_SUBTITLE || encoder.IsEncoder() == false {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not a subtitle encoding codec", codec)
	}

	self := &SubtitleProfile{
		SubtitleProfileMeta: SubtitleProfileMeta{
			Name:  encoder.Name(),
			codec: encoder,
			opts:  optionsForCodec(encoder),
		},
	}

	// Update internal codec parameters
	if err := self.setPar(); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r SubtitleProfile) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PROFILE INTERFACE

func (r SubtitleProfile) UUID() uuid.UUID {
	return r.Id
}

func (r SubtitleProfile) Type() CodecType {
	if r.codec == nil {
		return CodecType(ff.AVMEDIA_TYPE_UNKNOWN)
	}
	return CodecType(r.codec.Type())
}

func (r SubtitleProfile) Codec() *Codec {
	if r.codec == nil {
		return nil
	}
	return NewCodec(r.codec)
}

func (r SubtitleProfile) Par() *ff.AVCodecParameters {
	return types.Ptr(r.par)
}

// TimeBase always returns nil: unlike audio (sample rate) or video (frame
// rate), subtitle codecs have no rate-like field to derive one from, so the
// muxer/stream default applies.
func (r SubtitleProfile) TimeBase() *ff.AVRational {
	return nil
}

func (r SubtitleProfile) Options() json.RawMessage {
	return r.Opts
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - READER

// Expected column order: id, codec, opts.
func (r *SubtitleProfile) Scan(row pg.Row) error {
	if err := row.Scan(&r.Id, &r.Name, &r.Opts); err != nil {
		return err
	}

	// Set context and options
	encoder := ff.AVCodec_find_encoder_by_name(r.Name)
	if encoder == nil {
		return gomedia.ErrBadParameter.Withf("codec %q is not found", r.Name)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_SUBTITLE || encoder.IsEncoder() == false {
		return gomedia.ErrBadParameter.Withf("codec %q is not a subtitle encoding codec", r.Name)
	} else {
		r.codec = encoder
		r.opts = optionsForCodec(encoder)
	}

	// Set codec parameters
	if err := r.setPar(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - SELECTOR

func (r SubtitleProfileUUID) Select(bind *pg.Bind, op pg.Op) (string, error) {
	bind.Set("id", uuid.UUID(r))

	switch op {
	case pg.Get:
		return bind.Query("profile.subtitle_get"), nil
	case pg.Delete:
		return bind.Query("profile.subtitle_delete"), nil
	case pg.Update:
		return bind.Query("profile.subtitle_update"), nil
	default:
		return "", gomedia.ErrInternalError.Withf("unsupported SubtitleProfileUUID operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - WRITER

// Insert binds values and returns the insert query for a subtitle profile row.
func (r SubtitleProfileMeta) Insert(bind *pg.Bind) (string, error) {
	bind.Set("codec", r.Name)
	if r.Opts == nil {
		bind.Set("opts", map[string]any{})
	} else {
		bind.Set("opts", r.Opts)
	}
	return bind.Query("profile.subtitle_insert"), nil
}

// Update binds patch values for a subtitle profile row update.
func (r SubtitleProfileMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")

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

// Set a subtitle profile option. Every subtitle option is codec-specific
// (there's no universal field like bitrate or width), so this only ever
// touches the generic Opts dict. If value is nil, the option is removed.
func (r *SubtitleProfileMeta) Set(name string, value any) error {
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

	// Remove or set the option value
	if value == nil {
		delete(opts, name)
	} else if value, err := opt.Validate(value); err != nil {
		return err
	} else {
		opts[name] = value
	}

	// Set the option value in the options map
	if optsJSON, err := json.Marshal(opts); err != nil {
		return err
	} else {
		r.Opts = optsJSON
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - GET/SET OPTIONS

func (r *SubtitleProfileMeta) setPar() error {
	// Check for codec
	if r.codec == nil {
		return gomedia.ErrInternalError.With("codec is not set")
	}
	r.par.SetCodecType(r.codec.Type())
	r.par.SetCodecID(r.codec.ID())
	r.par.SetProfile(ff.AV_PROFILE_UNKNOWN)

	// Return success
	return nil
}
