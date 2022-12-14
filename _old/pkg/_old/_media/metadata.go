package media

import (
	"fmt"
	"strconv"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Metadata struct {
	dict *ffmpeg.AVDictionary
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMetadata(dict *ffmpeg.AVDictionary) *Metadata {
	return &Metadata{dict}
}

func (m *Metadata) Release() error {
	m.dict = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *Metadata) String() string {
	str := "<metadata"
	if keys := m.Keys(); len(keys) > 0 {
		str += " keys="
		for i, key := range keys {
			if i > 0 {
				str += ","
			}
			str += fmt.Sprint(key)
		}
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *Metadata) Keys() []MediaKey {
	if m.dict == nil {
		return nil
	}
	keys := make([]MediaKey, 0, m.dict.Count())
	entry := m.dict.Get("", nil, ffmpeg.AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		keys = append(keys, MediaKey(entry.Key()))
		entry = m.dict.Get("", entry, ffmpeg.AV_DICT_IGNORE_SUFFIX)
	}
	return keys
}

func (m *Metadata) Value(key MediaKey) interface{} {
	if entry := m.dict.Get(string(key), nil, ffmpeg.AV_DICT_IGNORE_SUFFIX); entry == nil {
		return nil
	} else if key == MEDIA_KEY_COMPILATION {
		if value, err := strconv.ParseInt(entry.Value(), 0, 32); err == nil {
			return value != 0
		} else {
			return nil
		}
	} else if key == MEDIA_KEY_GAPLESS_PLAYBACK {
		if value, err := strconv.ParseInt(entry.Value(), 0, 32); err == nil {
			return value != 0
		} else {
			return nil
		}
	} else if key == MEDIA_KEY_TRACK || key == MEDIA_KEY_DISC {
		n, _ := ParseTrackDisc(entry.Value())
		if n != 0 {
			return n
		} else {
			return nil
		}
	} else if key == MEDIA_KEY_YEAR {
		if value, err := strconv.ParseUint(entry.Value(), 0, 32); err == nil {
			return uint(value)
		} else {
			return nil
		}
	} else {
		return entry.Value()
	}
}
