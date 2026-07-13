package manager

import (
	"context"
	"time"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	chromaprintschema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioFingerprintLookup looks up an audio fingerprint from a fingerprint
func (m *Media) AudioFingerprintLookup(ctx context.Context, req schema.AudioFingerprintLookupRequest) (_ *schema.AudioFingerprintLookupResponse, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "AudioFingerprintLookup",
		attribute.Float64("duration", req.Duration),
		attribute.Int("metadata_count", len(req.Metadata)),
	)
	defer func() { endSpan(err) }()

	// Check for client
	if m.acoustIDClient == nil {
		return nil, gomedia.ErrNotImplemented.With("acoustid lookup client is not configured")
	}

	flags := metadataFlags(req.Metadata)
	matches, err := m.acoustIDClient.Lookup(ctx, req.Fingerprint, time.Duration(req.Duration*float64(time.Second)), flags)
	if err != nil {
		return nil, err
	}

	resp := &schema.AudioFingerprintLookupResponse{}
	if len(matches) > 0 {
		resp.Matches = [][]*chromaprintschema.ResponseMatch{matches}
	}

	return resp, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// metadataFlags converts metadata names to chromaprint lookup flags.
func metadataFlags(metadata []string) chromaprint.Meta {
	if len(metadata) == 0 {
		return chromaprint.META_ALL
	}

	var flags chromaprint.Meta
	for _, m := range metadata {
		switch m {
		case "recordings":
			flags |= chromaprint.META_RECORDING
		case "recordingids":
			flags |= chromaprint.META_RECORDINGID
		case "releases":
			flags |= chromaprint.META_RELEASE
		case "releaseids":
			flags |= chromaprint.META_RELEASEID
		case "releasegroups":
			flags |= chromaprint.META_RELEASEGROUP
		case "releasegroupids":
			flags |= chromaprint.META_RELEASEGROUPID
		case "tracks":
			flags |= chromaprint.META_TRACK
		case "compress":
			flags |= chromaprint.META_COMPRESS
		case "usermeta":
			flags |= chromaprint.META_USERMETA
		case "sources":
			flags |= chromaprint.META_SOURCE
		}
	}

	if flags == 0 {
		return chromaprint.META_ALL
	}

	return flags
}
