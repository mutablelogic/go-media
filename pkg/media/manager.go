package media

import (

	// Packages

	// Namespace imports

	. "github.com/djthorpe/go-errors"
	multierror "github.com/hashicorp/go-multierror"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type manager struct {
	media map[Media]bool
}

// Ensure manager complies with Manager interface
var _ Manager = (*manager)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *manager {
	m := new(manager)
	m.media = make(map[Media]bool, 10)
	return m
}

func (m *manager) Close() error {
	var result error

	// Close any opened media files
	for media := range m.media {
		if err := media.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Open media for reading and return it
func (m *manager) OpenFile(path string) (Media, error) {
	media, err := NewInputFile(path, func(media Media) error {
		if _, exists := m.media[media]; !exists {
			return ErrInternalAppError.With(media)
		}
		delete(m.media, media)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Add to map
	m.media[media] = true

	// Return success
	return media, nil
}

// Create media for writing and return it
func (m *manager) CreateFile(path string) (Media, error) {
	return nil, ErrNotImplemented
}
