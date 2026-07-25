package manager

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"time"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// SegmentAudio segments audio from an input reader and logs each segment.
// Segment sample payloads are discarded after logging.
func (m *Media) SegmentAudio(ctx context.Context, req schema.SegmentAudioRequest) error {
	if req.Reader == nil {
		return errors.New("input reader is nil")
	}
	outputDir := req.OutputDir
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	const sampleRate = 16000
	const defaultSilenceThreshold = 0.005
	var opts []segmenter.Opt

	if err := checkM4AEncoder(sampleRate); err != nil {
		return err
	}

	if req.Duration > 0 {
		opts = append(opts, segmenter.WithSegmentSize(req.Duration))
	}

	// Silence splitting is enabled explicitly, or implicitly when silence tuning flags are provided.
	useSilence := req.Silence || req.SilenceDuration > 0 || req.SilenceThreshold > 0
	if useSilence {
		threshold := req.SilenceThreshold
		if threshold <= 0 {
			threshold = defaultSilenceThreshold
		}

		switch {
		case req.SilenceThreshold > 0 && req.SilenceDuration > 0:
			opts = append(opts,
				segmenter.WithSilenceThreshold(threshold),
				segmenter.WithSilenceSize(req.SilenceDuration),
			)
		case req.SilenceThreshold > 0:
			opts = append(opts, segmenter.WithSilenceThreshold(threshold))
		case req.SilenceDuration > 0:
			opts = append(opts,
				segmenter.WithSilenceThreshold(threshold),
				segmenter.WithSilenceSize(req.SilenceDuration),
			)
		default:
			opts = append(opts, segmenter.WithSilenceThreshold(threshold))
		}
	}

	seg, err := segmenter.NewFromReader(req.Reader, sampleRate, opts...)
	if err != nil {
		return err
	}
	defer seg.Close()

	var (
		count        int
		totalSamples int
		segments     [][]int16
	)

	err = seg.DecodeInt16(ctx, func(ts time.Duration, samples []int16) error {
		count++
		sampleCount := len(samples)
		totalSamples += sampleCount

		segmentDuration := time.Duration(sampleCount) * time.Second / time.Duration(sampleRate)
		slog.InfoContext(ctx, "audio segment",
			"index", count,
			"timestamp", ts,
			"samples", sampleCount,
			"duration", segmentDuration,
		)

		copySamples := make([]int16, sampleCount)
		copy(copySamples, samples)
		segments = append(segments, copySamples)

		return nil
	})
	if err != nil {
		return err
	}

	totalSegments := len(segments)
	for i, samples := range segments {
		filename := fmt.Sprintf("%d-%d.m4a", i+1, totalSegments)
		path := filepath.Join(outputDir, filename)
		if err := encodeSegmentM4A(ctx, sampleRate, samples, path); err != nil {
			return err
		}
		slog.InfoContext(ctx, "audio segment encoded",
			"index", i+1,
			"segments", totalSegments,
			"path", path,
		)
	}

	totalDuration := time.Duration(totalSamples) * time.Second / time.Duration(sampleRate)
	slog.InfoContext(ctx, "audio segmentation complete",
		"segments", count,
		"samples", totalSamples,
		"duration", totalDuration,
	)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func encodeSegmentM4A(_ context.Context, sampleRate int, samples []int16, outPath string) error {
	audioPar, err := ffmpeg.NewAudioPar("fltp", "mono", sampleRate)
	if err != nil {
		return err
	}

	writer, err := ffmpeg.Create(outPath, ffmpeg.OptStream(0, audioPar))
	if err != nil {
		return err
	}
	defer writer.Close()

	const fallbackFrameSamples = 1152
	frameSamples := fallbackFrameSamples
	if stream := writer.Stream(0); stream != nil && stream.FrameSize() > 0 {
		frameSamples = stream.FrameSize()
	}

	frame, err := ffmpeg.NewFrame(audioPar)
	if err != nil {
		return err
	}
	defer frame.Close()

	pts := 0
	for offset := 0; offset < len(samples); offset += frameSamples {
		chunkEnd := offset + frameSamples
		if chunkEnd > len(samples) {
			chunkEnd = len(samples)
		}

		chunk := make([]float32, frameSamples)
		for i, sample := range samples[offset:chunkEnd] {
			chunk[i] = float32(sample) / float32(math.MaxInt16)
		}

		frame.SetPts(int64(pts))
		if err := frame.SetFloat32(0, chunk); err != nil {
			return err
		}
		if err := writer.EncodeFrame(0, frame); err != nil {
			return err
		}
		pts += frameSamples
	}

	if err := writer.EncodeFrame(0, nil); err != nil {
		return err
	}

	return nil
}

func checkM4AEncoder(sampleRate int) error {
	audioPar, err := ffmpeg.NewAudioPar("fltp", "mono", sampleRate)
	if err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "gomedia-m4a-check-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	probePath := filepath.Join(dir, "probe.m4a")
	writer, err := ffmpeg.Create(probePath, ffmpeg.OptStream(0, audioPar))
	if err != nil {
		return fmt.Errorf("m4a/aac encoding is unavailable in this build: %w; check available encoders with 'gomedia codecs --type audio --is-encoder --name aac'", err)
	}

	return writer.Close()
}
