//go:build chromaprint

package manager

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioFingerprint generates an audio fingerprint from an input file or reader.
func (m *Media) AudioFingerprint(ctx context.Context, req schema.AudioFingerprintRequest) (_ *schema.AudioFingerprintResponse, err error) {
	// Determine input source and open file if needed.
	var reader io.Reader
	if req.Input != "" {
		f, err := os.Open(req.Input)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		reader = f
	} else if req.Reader != nil {
		reader = req.Reader
	} else {
		return nil, fmt.Errorf("either Reader or Input must be set")
	}

	// Convert duration in seconds.
	var dur time.Duration
	if req.Duration > 0 {
		dur = time.Duration(req.Duration * float64(time.Second))
	}

	// Build segmenter options for explicit input format/options.
	var opts []segmenter.Opt
	if req.InputFormat != "" || len(req.InputOpts) > 0 {
		opts = append(opts, segmenter.WithFFmpegOpt(ffmpeg.WithInput(req.InputFormat, req.InputOpts...)))
	}

	// Generate fingerprint.
	fpResult, err := chromaprint.Fingerprint(ctx, reader, dur, opts...)
	if err != nil {
		return nil, err
	}

	return &schema.AudioFingerprintResponse{
		Fingerprint: fpResult.Fingerprint,
		Duration:    fpResult.Duration,
	}, nil
}
