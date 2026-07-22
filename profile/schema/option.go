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

type Option struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Default     any      `json:"default,omitempty"`
	Const       []Option `json:"const,omitempty"`
	Min         any      `json:"min,omitempty"`
	Max         any      `json:"max,omitempty"`
	Unit        string   `json:"unit,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewOption(ctx *ff.AVOption) Option {
	if ctx == nil {
		return Option{}
	}

	option := Option{
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
			option.Min = min
		}
		if max := ctx.Max(); !math.IsNaN(max) && !math.IsInf(max, 0) {
			option.Max = max
		}
	}

	return option
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
