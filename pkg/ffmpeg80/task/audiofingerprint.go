package task

import (
	"context"
	"os"
	"time"

	// Packages
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg80"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg80/schema"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioFingerprint generates an audio fingerprint and optionally performs AcoustID lookup
func (m *Manager) AudioFingerprint(ctx context.Context, req *schema.AudioFingerprintRequest) (*schema.AudioFingerprintResponse, error) {
	// Build segmenter options from Request.Path if it contains format info
	var opts []segmenter.Opt
	if req.Path != "" && req.Reader != nil {
		// When Reader is set and Path is set, Path is interpreted as the format
		opts = append(opts, segmenter.WithFFmpegOpt(ffmpeg.WithInput(req.Path)))
	}

	// Open the reader
	var reader *os.File
	if req.Reader == nil {
		// Open from path
		f, err := os.Open(req.Path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		reader = f
	}

	// Convert duration
	var dur time.Duration
	if req.Duration > 0 {
		dur = time.Duration(req.Duration * float64(time.Second))
	}

	// If lookup is requested, we need a client
	if req.Lookup {
		// Get API key from manager or environment
		apiKey := m.chromaprintKey
		if apiKey == "" {
			apiKey = os.Getenv("CHROMAPRINT_KEY")
		}

		// Create chromaprint client
		client, err := chromaprint.NewClient(apiKey)
		if err != nil {
			return nil, err
		}

		// Build metadata flags
		flags := metadataFlags(req.Metadata)

		// Perform match with lookup (this also generates the fingerprint)
		var matches []*chromaprint.ResponseMatch
		if req.Reader != nil {
			matches, err = client.Match(ctx, req.Reader, dur, flags, opts...)
		} else {
			matches, err = client.Match(ctx, reader, dur, flags, opts...)
		}
		if err != nil {
			return nil, err
		}

		// We need to generate the fingerprint to include it in the response
		// Reopen the file/reader since it was consumed by Match
		var fpResult *chromaprint.FingerprintResult
		if req.Reader == nil {
			// Reopen file for fingerprinting
			f2, err := os.Open(req.Path)
			if err != nil {
				return nil, err
			}
			defer f2.Close()
			fpResult, err = chromaprint.Fingerprint(ctx, f2, dur, opts...)
			if err != nil {
				return nil, err
			}
		} else {
			// Reader was already consumed, we can't get the fingerprint again
			// This is a limitation when using Reader mode with Lookup
			fpResult = &chromaprint.FingerprintResult{
				Fingerprint: "", // Cannot regenerate from consumed reader
				Duration:    dur,
			}
		}

		// Build response
		resp := &schema.AudioFingerprintResponse{
			Fingerprint: fpResult.Fingerprint,
			Duration:    fpResult.Duration.Seconds(),
		}

		// Convert matches
		if len(matches) > 0 {
			resp.Matches = make([]chromaprint.ResponseMatch, len(matches))
			for i, m := range matches {
				resp.Matches[i] = *m
			}
		}

		return resp, nil
	}

	// Just fingerprint, no lookup
	var result *chromaprint.FingerprintResult
	var err error
	if req.Reader != nil {
		result, err = chromaprint.Fingerprint(ctx, req.Reader, dur, opts...)
	} else {
		result, err = chromaprint.Fingerprint(ctx, reader, dur, opts...)
	}
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
