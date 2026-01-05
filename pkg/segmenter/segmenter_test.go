package segmenter_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func testDataPath(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "..", "..", "etc", "test")
}

func TestSegmenter_New(t *testing.T) {
	assert := assert.New(t)

	path := filepath.Join(testDataPath(t), "sample.mp3")
	seg, err := segmenter.New(path, 16000)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer seg.Close()

	assert.NotNil(seg)
	assert.Equal(16000, seg.SampleRate())
	assert.Greater(seg.Duration(), time.Duration(0))
	t.Logf("Duration: %v", seg.Duration())
}

func TestSegmenter_NewFromReader(t *testing.T) {
	assert := assert.New(t)

	path := filepath.Join(testDataPath(t), "sample.mp3")
	f, err := os.Open(path)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer f.Close()

	seg, err := segmenter.NewFromReader(f, 16000)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer seg.Close()

	assert.NotNil(seg)
	assert.Equal(16000, seg.SampleRate())
}

func TestSegmenter_DecodeFloat32_NoSegmentation(t *testing.T) {
	assert := assert.New(t)

	path := filepath.Join(testDataPath(t), "sample.mp3")
	seg, err := segmenter.New(path, 16000)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer seg.Close()

	var totalSamples int
	var segmentCount int

	err = seg.DecodeFloat32(context.Background(), func(ts time.Duration, samples []float32) error {
		segmentCount++
		totalSamples += len(samples)
		t.Logf("Segment %d at %v: %d samples", segmentCount, ts, len(samples))
		return nil
	})

	assert.NoError(err)
	assert.Greater(totalSamples, 0)
	t.Logf("Total: %d segments, %d samples", segmentCount, totalSamples)
}

func TestSegmenter_DecodeFloat32_WithSegmentSize(t *testing.T) {
	assert := assert.New(t)

	path := filepath.Join(testDataPath(t), "sample.mp3")
	seg, err := segmenter.New(path, 16000, segmenter.WithSegmentSize(200*time.Millisecond))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer seg.Close()

	var totalSamples int
	var segmentCount int
	expectedSamplesPerSegment := 16000 * 200 / 1000 // 3200 samples

	err = seg.DecodeFloat32(context.Background(), func(ts time.Duration, samples []float32) error {
		segmentCount++
		totalSamples += len(samples)
		t.Logf("Segment %d at %v: %d samples", segmentCount, ts, len(samples))
		return nil
	})

	assert.NoError(err)
	assert.Greater(segmentCount, 1)
	t.Logf("Total: %d segments, %d samples (expected ~%d per segment)", segmentCount, totalSamples, expectedSamplesPerSegment)
}

func TestSegmenter_DecodeInt16_WithSegmentSize(t *testing.T) {
	assert := assert.New(t)

	path := filepath.Join(testDataPath(t), "sample.mp3")
	seg, err := segmenter.New(path, 16000, segmenter.WithSegmentSize(200*time.Millisecond))
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer seg.Close()

	var totalSamples int
	var segmentCount int

	err = seg.DecodeInt16(context.Background(), func(ts time.Duration, samples []int16) error {
		segmentCount++
		totalSamples += len(samples)
		t.Logf("Segment %d at %v: %d samples", segmentCount, ts, len(samples))
		return nil
	})

	assert.NoError(err)
	assert.Greater(segmentCount, 1)
	t.Logf("Total: %d segments, %d samples", segmentCount, totalSamples)
}

func TestSegmenter_WithSilenceDetection(t *testing.T) {
	assert := assert.New(t)

	// Use the JFK WAV file which has speech with pauses
	path := filepath.Join(testDataPath(t), "jfk.wav")
	seg, err := segmenter.New(path, 16000,
		segmenter.WithSegmentSize(5*time.Second),
		segmenter.WithDefaultSilence(),
	)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer seg.Close()

	var segmentCount int
	var totalDuration time.Duration

	err = seg.DecodeFloat32(context.Background(), func(ts time.Duration, samples []float32) error {
		segmentCount++
		duration := time.Duration(len(samples)) * time.Second / time.Duration(16000)
		totalDuration += duration
		t.Logf("Segment %d at %v: %d samples (%.2fs)", segmentCount, ts, len(samples), duration.Seconds())
		return nil
	})

	assert.NoError(err)
	assert.Greater(segmentCount, 0)
	t.Logf("Total: %d segments, %.2fs", segmentCount, totalDuration.Seconds())
}

func TestSegmenter_InvalidSampleRate(t *testing.T) {
	path := filepath.Join(testDataPath(t), "sample.mp3")

	_, err := segmenter.New(path, 0)
	assert.Error(t, err)

	_, err = segmenter.New(path, -1)
	assert.Error(t, err)
}

func TestSegmenter_InvalidSegmentSize(t *testing.T) {
	path := filepath.Join(testDataPath(t), "sample.mp3")

	_, err := segmenter.New(path, 16000, segmenter.WithSegmentSize(10*time.Millisecond))
	assert.Error(t, err) // Too short
}

func TestSegmenter_FileNotFound(t *testing.T) {
	_, err := segmenter.New("/nonexistent/file.mp3", 16000)
	assert.Error(t, err)
}

func TestSegmenter_NilCallback(t *testing.T) {
	path := filepath.Join(testDataPath(t), "sample.mp3")
	seg, err := segmenter.New(path, 16000)
	require.NoError(t, err)
	defer seg.Close()

	err = seg.DecodeFloat32(context.Background(), nil)
	assert.Error(t, err)

	err = seg.DecodeInt16(context.Background(), nil)
	assert.Error(t, err)
}
