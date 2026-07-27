//go:build !chromaprint

package manager

import (
	"context"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioFingerprint is unavailable unless built with -tags chromaprint.
func (m *Media) AudioFingerprint(_ context.Context, _ schema.AudioFingerprintRequest) (_ *schema.AudioFingerprintResponse, err error) {
	return nil, gomedia.ErrNotImplemented.With("audio fingerprint is unavailable: build with -tags chromaprint")
}
