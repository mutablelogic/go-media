package manager

import (
	"context"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioFingerprintLookup looks up an audio fingerprint from a fingerprint
func (m *Media) AudioFingerprintLookup(ctx context.Context, req schema.AudioFingerprintLookupRequest) (_ schema.AudioFingerprintLookupResponse, err error) {
	return schema.AudioFingerprintLookupResponse{}, gomedia.ErrNotImplemented.With("audio fingerprint lookup is not implemented")
}
