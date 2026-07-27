package schema

import (
	"fmt"
	"math"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Option.Type and OptionConst.Type hold one of the strings
// ff.AVOptionType.String() can produce (see the enum tag on both fields).

type Option struct {
	Name        string        `json:"name,omitempty" help:"Option name." example:"b"`
	Description string        `json:"description,omitempty" help:"Human-readable option description." example:"Set bitrate in bits/s."`
	Type        string        `json:"type,omitempty" help:"Option value type." example:"int" enum:"flags,int,int64,uint,uint64,double,float,string,rational,binary,dict,const,image_size,pixel_fmt,sample_fmt,video_rate,duration,color,bool,chlayout,flag_array,unknown"`
	Default     any           `json:"default,omitempty" help:"Default value." example:"128000"`
	Const       []OptionConst `json:"const,omitempty" help:"Fixed set of allowed values, if this option only accepts specific values."`
	Min         any           `json:"min,omitempty" help:"Minimum allowed value, if applicable." example:"0"`
	Max         any           `json:"max,omitempty" help:"Maximum allowed value, if applicable." example:"192000"`
	Unit        string        `json:"unit,omitempty" help:"Unit the value is expressed in, if applicable." example:"bps"`
}

// OptionConst describes one of an Option's fixed set of allowed values (an
// AV_OPT_TYPE_CONST entry, or a hand-built constant like a supported sample
// rate, pixel format or codec profile). It's a separate, non-recursive type
// from Option - rather than reusing Option itself - because a const entry
// is always terminal: it never has its own set of constants. Option.Const
// being []Option would make Option a self-referential type, which the
// JSONSchema generator can't represent.
type OptionConst struct {
	Name        string `json:"name,omitempty" help:"Symbolic name for this value, if any." example:"fast"`
	Description string `json:"description,omitempty" help:"Human-readable description of this value."`
	Type        string `json:"type,omitempty" help:"Value type." example:"int" enum:"flags,int,int64,uint,uint64,double,float,string,rational,binary,dict,const,image_size,pixel_fmt,sample_fmt,video_rate,duration,color,bool,chlayout,flag_array,unknown"`
	Default     any    `json:"default,omitempty" help:"The underlying value." example:"1"`
	Min         any    `json:"min,omitempty" help:"Minimum allowed value, if applicable."`
	Max         any    `json:"max,omitempty" help:"Maximum allowed value, if applicable."`
	Unit        string `json:"unit,omitempty" help:"Unit the value is expressed in, if applicable."`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewOption(ctx *ff.AVOption) Option {
	if ctx == nil {
		return Option{}
	}
	c := newOptionConst(ctx)
	return Option{
		Name:        c.Name,
		Description: c.Description,
		Type:        c.Type,
		Default:     c.Default,
		Min:         c.Min,
		Max:         c.Max,
		Unit:        c.Unit,
	}
}

// NewOptionConst builds a terminal OptionConst (e.g. an AV_OPT_TYPE_CONST
// entry) from an FFmpeg option. Use this instead of NewOption when building
// an Option's Const list.
func NewOptionConst(ctx *ff.AVOption) OptionConst {
	if ctx == nil {
		return OptionConst{}
	}
	return newOptionConst(ctx)
}

func newOptionConst(ctx *ff.AVOption) OptionConst {
	c := OptionConst{
		Name:        ctx.Name(),
		Type:        ctx.Type().String(),
		Default:     normalizeOptionValue(ctx.DefaultVal(), ctx.Type()),
		Unit:        ctx.Unit(),
		Description: ctx.Help(),
	}

	switch ctx.Type() {
	case ff.AV_OPT_TYPE_INT, ff.AV_OPT_TYPE_INT64, ff.AV_OPT_TYPE_UINT, ff.AV_OPT_TYPE_UINT64,
		ff.AV_OPT_TYPE_DOUBLE, ff.AV_OPT_TYPE_FLOAT, ff.AV_OPT_TYPE_DURATION:
		if min := ctx.Min(); !math.IsNaN(min) && !math.IsInf(min, 0) {
			c.Min = min
		}
		if max := ctx.Max(); !math.IsNaN(max) && !math.IsInf(max, 0) {
			c.Max = max
		}
	}

	return c
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o Option) String() string {
	return types.Stringify(o)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (o Option) Validate(value any) (any, error) {
	if value == nil {
		return nil, gomedia.ErrBadParameter.Withf("option %q cannot be nil", o.Name)
	}

	// Check against constants. Consts derived from ffmpeg's own AVOption consts
	// carry a symbolic Name (e.g. "fast") with the underlying value in Default;
	// consts built by hand (sample_rate, sample_format, channel_layout in
	// codec.go) have no Name at all and are selected by Default directly. Accept
	// either form.
	if len(o.Const) > 0 {
		for _, v := range o.Const {
			if optionValueEqual(value, v.Name) || optionValueEqual(value, v.Default) {
				return value, nil
			}
		}
		return nil, fmt.Errorf("option %q must be one of %v", o.Name, o.Const)
	}

	switch o.Type {
	case "bool":
		if _, ok := value.(bool); !ok {
			return nil, fmt.Errorf("option %q must be a bool", o.Name)
		}
	case "int", "int64", "uint", "uint64", "double", "float", "duration":
		value, ok := numericValue(value)
		if !ok {
			return nil, fmt.Errorf("option %q must be numeric", o.Name)
		}
		if min, ok := numericValue(o.Min); ok && value < min {
			return nil, fmt.Errorf("option %q must be >= %v", o.Name, o.Min)
		}
		if max, ok := numericValue(o.Max); ok && value > max {
			return nil, fmt.Errorf("option %q must be <= %v", o.Name, o.Max)
		}
	case "string", "image_size", "pixel_fmt", "sample_fmt", "video_rate", "color", "chlayout", "binary", "dict":
		if _, ok := value.(string); !ok {
			return nil, fmt.Errorf("option %q must be a string", o.Name)
		}
	case "rational":
		if _, ok := value.(ff.AVRational); !ok {
			return nil, fmt.Errorf("option %q must be an AVRational", o.Name)
		}
	}

	return value, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func normalizeOptionValue(value any, t ff.AVOptionType) any {
	switch v := value.(type) {
	case nil:
		return nil
	case ff.AVRational:
		return v
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil
		}
		return v
	case float32:
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil
		}
		return v
	default:
		switch t {
		case ff.AV_OPT_TYPE_BOOL:
			if b, ok := value.(bool); ok {
				return b
			}
		}
		return value
	}
}

func optionValueEqual(a, b any) bool {
	if na, ok := numericValue(a); ok {
		if nb, ok := numericValue(b); ok {
			return na == nb
		}
	}
	return a == b
}

func numericValue(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
