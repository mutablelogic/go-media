package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	// Packages
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	schema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	chromaprint *chromaprint.Client
}

type Opt func(*Manager) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(opts ...Opt) (*Manager, error) {
	m := &Manager{}
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// WithChromaprintKey sets the AcoustID API key for lookups
func WithChromaprintKey(key string) Opt {
	return func(m *Manager) error {
		client, err := chromaprint.NewClient(key)
		if err != nil {
			return err
		}
		m.chromaprint = client
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioFingerprint generates an audio fingerprint and optionally performs AcoustID lookup
func (m *Manager) AudioFingerprint(ctx context.Context, req *schema.AudioFingerprintRequest) (*schema.AudioFingerprintResponse, error) {
	// Ensure chromaprint client is available if lookup is requested
	if req.Lookup && m.chromaprint == nil {
		return nil, fmt.Errorf("lookup requested but no AcoustID API key configured")
	}
	// Build segmenter options from Request if needed
	var opts []segmenter.Opt

	// Determine input source
	var inputPath string
	var reader io.Reader

	if req.Path != "" {
		inputPath = req.Path
	} else if req.Reader != nil {
		reader = req.Reader
	} else {
		return nil, fmt.Errorf("either Reader or Path must be set")
	}

	// Convert duration
	var dur time.Duration
	if req.Duration > 0 {
		dur = time.Duration(req.Duration * float64(time.Second))
	}

	// If lookup is requested, we need a client
	if req.Lookup {
		// Build metadata flags
		flags := metadataFlags(req.Metadata)

		// Perform match with lookup (using path or reader)
		var matches []*chromaprint.ResponseMatch

		if inputPath != "" {
			// Open file for matching
			f, err := os.Open(inputPath)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			// Generate fingerprint
			fpResult, err := chromaprint.Fingerprint(ctx, f, dur, opts...)
			if err != nil {
				return nil, err
			}

			// Lookup matches
			matches, err = m.chromaprint.Lookup(fpResult.Fingerprint, time.Duration(fpResult.Duration*float64(time.Second)), flags)
			if err != nil {
				return nil, err
			}

			// Re-open for fingerprint
			f2, err := os.Open(inputPath)
			if err != nil {
				return nil, err
			}
			defer f2.Close()

			fpResult, err := chromaprint.Fingerprint(ctx, f2, dur, opts...)
			if err != nil {
				return nil, err
			}

			// Build response
			resp := &schema.AudioFingerprintResponse{
				Fingerprint: fpResult.Fingerprint,
				Duration:    fpResult.Duration.Seconds(),
			}

			// Add matches directly
			if len(matches) > 0 {
				resp.Matches = [][]*chromaprint.ResponseMatch{matches}
			}

			return resp, nil
		} else {
			// Using reader - can't re-read for lookup
			return nil, fmt.Errorf("lookup requires re-reading the file; use Path instead of Reader")
		}
	}

	// Just fingerprint, no lookup
	if inputPath != "" {
		f, err := os.Open(inputPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		reader = f
	}

	result, err := chromaprint.Fingerprint(ctx, reader, dur, opts...)
	if err != nil {
		return nil, err
	}

	return &schema.AudioFingerprintResponse{
		Fingerprint: result.Fingerprint,
		Duration:    result.Duration.Seconds(),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// metadataFlags converts string metadata flags to chromaprint.Meta flags
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

	return flags
}
