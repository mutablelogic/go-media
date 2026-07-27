package manager

import (
	"context"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	testcard "github.com/mutablelogic/go-media/gomedia/testcard"
	profile "github.com/mutablelogic/go-media/profile/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// CreateSource creates a new test card source with the given name and audio
// profile, and registers it under that name. The name must not already be
// in use.
func (m *Media) CreateSource(ctx context.Context, name string, audio *profile.AudioProfile) (schema.Source, error) {
	if name == "" {
		return nil, gomedia.ErrBadParameter.With("source name must not be empty")
	}

	src, err := testcard.New(audio)
	if err != nil {
		return nil, err
	}

	m.sourcesMu.Lock()
	defer m.sourcesMu.Unlock()

	if m.sources == nil {
		m.sources = make(map[string]schema.Source)
	} else if _, exists := m.sources[name]; exists {
		src.Close()
		return nil, gomedia.ErrBadParameter.Withf("source %q already exists", name)
	}

	m.sources[name] = src
	return src, nil
}

// GetSource returns the source registered under name.
func (m *Media) GetSource(ctx context.Context, name string) (schema.Source, error) {
	m.sourcesMu.RLock()
	defer m.sourcesMu.RUnlock()

	src, exists := m.sources[name]
	if !exists {
		return nil, gomedia.ErrNotFound.Withf("source %q not found", name)
	}
	return src, nil
}

// DeleteSource removes and closes the source registered under name.
func (m *Media) DeleteSource(ctx context.Context, name string) error {
	m.sourcesMu.Lock()
	defer m.sourcesMu.Unlock()

	src, exists := m.sources[name]
	if !exists {
		return gomedia.ErrNotFound.Withf("source %q not found", name)
	}

	delete(m.sources, name)
	return src.Close()
}
