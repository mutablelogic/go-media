package media

import (
	// Packages

	"fmt"
	"regexp"
	"strconv"
	"time"

	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type metadata struct {
	ctx *ffmpeg.AVDictionary
}

// Ensure manager complies with Manager interface
var _ Metadata = (*metadata)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// 1/2 etc
	reTrackDisc = regexp.MustCompile(`^(\\d+)/(\\d+)$`)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMetadata(ctx *ffmpeg.AVDictionary) *metadata {
	if ctx == nil {
		return nil
	}
	return &metadata{
		ctx: ctx,
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (metadata *metadata) String() string {
	str := "<media.metadata"
	for _, key := range metadata.Keys() {
		switch v := metadata.Value(key).(type) {
		case string:
			str += fmt.Sprintf(" %s=%q", key, v)
		default:
			if v == nil {
				str += fmt.Sprintf(" %s=nil", key)
			} else {
				str += fmt.Sprintf(" %s=%v", key, v)
			}
		}
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (metadata *metadata) Keys() []MediaKey {
	if metadata.ctx == nil {
		return nil
	}
	count := ffmpeg.AVUtil_av_dict_count(metadata.ctx)
	result := make([]MediaKey, 0, count)
	for _, key := range ffmpeg.AVUtil_av_dict_keys(metadata.ctx) {
		result = append(result, MediaKey(key))
	}
	return result
}

func (metadata *metadata) Value(key MediaKey) any {
	if metadata.ctx == nil {
		return nil
	}
	entry := ffmpeg.AVUtil_av_dict_get(metadata.ctx, string(key), nil, ffmpeg.AV_DICT_IGNORE_SUFFIX)
	if entry == nil {
		return nil
	}
	switch key {
	case MEDIA_KEY_COMPILATION, MEDIA_KEY_GAPLESS_PLAYBACK: // int -> bool
		if value, err := strconv.ParseInt(entry.Value(), 0, 32); err == nil {
			return value != 0
		} else if bool, err := strconv.ParseBool(entry.Value()); err == nil {
			return bool
		} else {
			return nil
		}
	case MEDIA_KEY_TRACK, MEDIA_KEY_DISC: // ddd/ddd -> []uint
		if x, y, err := parseTrackDisc(entry.Value()); err == nil {
			return []uint{x, y}
		} else {
			return nil
		}
	case MEDIA_KEY_YEAR: // uint
		if value, err := strconv.ParseUint(entry.Value(), 0, 32); err == nil {
			return uint(value)
		} else {
			return nil
		}
	case MEDIA_KEY_CREATED: // date.time
		if t, err := time.Parse(time.RFC3339, entry.Value()); err == nil {
			if t.IsZero() || t.Unix() == 0 {
				return nil
			} else {
				return t
			}
		} else {
			return nil
		}
	default:
		return entry.Value()
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func parseTrackDisc(value string) (uint, uint, error) {
	// parse d/d
	if nm := reTrackDisc.FindStringSubmatch(value); len(nm) == 3 {
		if n, err := strconv.ParseUint(nm[1], 0, 64); err != nil {
			return 0, 0, err
		} else if m, err := strconv.ParseUint(nm[2], 0, 64); err != nil {
			return 0, 0, err
		} else {
			return uint(n), uint(m), nil
		}
	} else if n, err := strconv.ParseUint(value, 0, 64); err != nil {
		return 0, 0, err
	} else {
		return uint(n), 0, nil
	}
}
