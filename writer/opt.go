package writer

import (
	// Packages
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	streams  map[int]profile.Profile
	metadata []gomedia.Metadata
}

type Opt func(o *opts) error

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (o *opts) apply(opts ...Opt) error {
	o.streams = make(map[int]profile.Profile, len(opts))
	for _, optFunc := range opts {
		if err := optFunc(o); err != nil {
			return err
		}
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Append metadata to the output file, including artwork
func WithMetadata(meta ...gomedia.Metadata) Opt {
	return func(o *opts) error {
		o.metadata = append(o.metadata, meta...)
		return nil
	}
}

// Append stream profile to the output file. If the stream index is 0, the next
// available stream index will be used.
func WithProfile(stream int, profile profile.Profile) Opt {
	return func(o *opts) error {
		if stream < 0 {
			return gomedia.ErrBadParameter.Withf("stream index must be non-negative")
		} else if stream == 0 {
			for stream = 0; ; stream++ {
				if _, exists := o.streams[stream]; !exists {
					break
				}
			}
		}
		if profile == nil {
			return gomedia.ErrBadParameter.Withf("profile must be non-nil")
		}
		if _, exists := o.streams[stream]; exists {
			return gomedia.ErrBadParameter.Withf("stream %d already has a profile", stream)
		}
		o.streams[stream] = profile
		return nil
	}
}
