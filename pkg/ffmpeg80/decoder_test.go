package ffmpeg_test

import (
	"context"
	"io"
	"testing"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg80"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

const (
	TEST_MP4 = "../../etc/test/sample.mp4"
	TEST_MP3 = "../../etc/test/sample.mp3"
	TEST_WAV = "../../etc/test/jfk.wav"
	TEST_JPG = "../../etc/test/sample.jpg"
	TEST_PNG = "../../etc/test/sample.png"
)

func Test_Decode_001(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Decode all packets
	ctx := context.Background()
	packetCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		assert.NotNil(pkt)
		assert.GreaterOrEqual(stream, 0)
		packetCount++
		return nil
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Decoded %d packets from %s", packetCount, TEST_MP4)
}

func Test_Decode_002(t *testing.T) {
	assert := assert.New(t)

	// Open audio file
	reader, err := ffmpeg.Open(TEST_MP3)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Decode all packets and verify stream indices
	ctx := context.Background()
	packetCount := 0
	streamSet := make(map[int]bool)
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		assert.NotNil(pkt)
		assert.GreaterOrEqual(stream, 0)
		assert.Equal(stream, pkt.Stream())
		streamSet[stream] = true
		packetCount++
		return nil
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	assert.Greater(len(streamSet), 0)
	t.Logf("Decoded %d packets from %d streams in %s", packetCount, len(streamSet), TEST_MP3)
}

func Test_Decode_003(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Count packets per stream
	ctx := context.Background()
	streamPackets := make(map[int]int)
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		streamPackets[stream]++
		return nil
	})

	assert.NoError(err)
	assert.Greater(len(streamPackets), 0)

	for stream, count := range streamPackets {
		t.Logf("Stream %d: %d packets", stream, count)
		assert.Greater(count, 0)
	}
}

func Test_Decode_004(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Stop after 10 packets using io.EOF
	ctx := context.Background()
	packetCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		packetCount++
		if packetCount >= 10 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.Equal(10, packetCount)
	t.Logf("Stopped after %d packets", packetCount)
}

func Test_Decode_005(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Cancel context after short duration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	packetCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		packetCount++
		// Slow down to ensure timeout is hit
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	assert.Error(err)
	assert.Equal(context.DeadlineExceeded, err)
	t.Logf("Decoded %d packets before timeout", packetCount)
}

func Test_Decode_006(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before decode

	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		assert.Fail("Should not be called")
		return nil
	})

	assert.Error(err)
	assert.Equal(context.Canceled, err)
}

func Test_Decode_007(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Verify packet properties
	ctx := context.Background()
	foundValidPts := false
	foundData := false
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		// Check for valid timestamps
		if pkt.Pts() >= 0 {
			foundValidPts = true
		}
		// Check for packet data
		if len(pkt.Bytes()) > 0 {
			foundData = true
		}
		if foundValidPts && foundData {
			return io.EOF // Found what we need
		}
		return nil
	})

	assert.NoError(err)
	assert.True(foundValidPts, "Should find packets with valid PTS")
	assert.True(foundData, "Should find packets with data")
}

func Test_Decode_008(t *testing.T) {
	assert := assert.New(t)

	// Open audio file
	reader, err := ffmpeg.Open(TEST_WAV)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Decode and check packet sizes
	ctx := context.Background()
	totalBytes := 0
	packetCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		size := pkt.Size()
		assert.GreaterOrEqual(size, 0)
		totalBytes += size
		packetCount++
		return nil
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	assert.Greater(totalBytes, 0)
	t.Logf("Decoded %d packets, %d total bytes", packetCount, totalBytes)
}

func Test_Decode_009(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)

	// Close reader
	err = reader.Close()
	assert.NoError(err)

	// Try to decode after close
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		assert.Fail("Should not be called")
		return nil
	})

	assert.Error(err)
	assert.Contains(err.Error(), "closed")
}

func Test_Decode_010(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Get expected streams
	streams := reader.Streams(media.ANY)
	assert.NotEmpty(streams)

	// Track which streams we see
	seenStreams := make(map[int]bool)
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		seenStreams[stream] = true
		return nil
	})

	assert.NoError(err)
	assert.NotEmpty(seenStreams)
	t.Logf("Expected %d streams, saw packets from %d streams", len(streams), len(seenStreams))
}

func Test_Decode_011(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// First decode - count packets
	ctx := context.Background()
	firstCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		firstCount++
		return nil
	})
	assert.NoError(err)
	assert.Greater(firstCount, 0)

	// Seek back to start
	streams := reader.Streams(media.ANY)
	if len(streams) > 0 {
		err = reader.Seek(streams[0].Index(), 0)
		assert.NoError(err)
	}

	// Second decode should work and get similar count
	secondCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		secondCount++
		return nil
	})
	assert.NoError(err)
	assert.Greater(secondCount, 0)

	t.Logf("First decode: %d packets, second decode: %d packets", firstCount, secondCount)
	// Counts should be similar after seek
	assert.InDelta(firstCount, secondCount, float64(firstCount)*0.1) // Within 10%
}

func Test_Decode_012(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Verify timestamps are reasonable
	ctx := context.Background()
	lastPts := make(map[int]int64)
	nonDecreasingCount := 0
	totalPackets := 0

	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		pts := pkt.Pts()
		if lastPts[stream] <= pts {
			nonDecreasingCount++
		}
		lastPts[stream] = pts
		totalPackets++
		return nil
	})

	assert.NoError(err)
	assert.Greater(totalPackets, 0)
	// Most packets should have non-decreasing PTS within their stream
	t.Logf("Non-decreasing PTS: %d/%d packets", nonDecreasingCount, totalPackets)
}

////////////////////////////////////////////////////////////////////////////////
// DEMUX TESTS

func Test_Demux_001(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Demux all streams without mapping (decode all)
	ctx := context.Background()
	frameCount := 0
	err = reader.Demux(ctx, nil, func(stream int, frame *ffmpeg.Frame) error {
		assert.NotNil(frame)
		assert.GreaterOrEqual(stream, 0)
		frameCount++
		return nil
	})

	assert.NoError(err)
	assert.Greater(frameCount, 0)
	t.Logf("Decoded %d frames from %s", frameCount, TEST_MP4)
}

func Test_Demux_002(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Count frames per stream
	ctx := context.Background()
	streamFrames := make(map[int]int)
	err = reader.Demux(ctx, nil, func(stream int, frame *ffmpeg.Frame) error {
		streamFrames[stream]++
		return nil
	})

	assert.NoError(err)
	assert.Greater(len(streamFrames), 0)

	for stream, count := range streamFrames {
		t.Logf("Stream %d: %d frames", stream, count)
		assert.Greater(count, 0)
	}
}

func Test_Demux_003(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Map only video streams
	ctx := context.Background()
	frameCount := 0
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.VIDEO {
			return par, nil // Decode video
		}
		return nil, nil // Ignore others
	}, func(stream int, frame *ffmpeg.Frame) error {
		assert.Equal(media.VIDEO, frame.Type())
		frameCount++
		return nil
	})

	assert.NoError(err)
	assert.Greater(frameCount, 0)
	t.Logf("Decoded %d video frames", frameCount)
}

func Test_Demux_004(t *testing.T) {
	assert := assert.New(t)

	// Open audio file
	reader, err := ffmpeg.Open(TEST_MP3)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Map only audio streams
	ctx := context.Background()
	frameCount := 0
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.AUDIO {
			return par, nil // Decode audio
		}
		return nil, nil // Ignore others
	}, func(stream int, frame *ffmpeg.Frame) error {
		assert.Equal(media.AUDIO, frame.Type())
		frameCount++
		return nil
	})

	assert.NoError(err)
	assert.Greater(frameCount, 0)
	t.Logf("Decoded %d audio frames", frameCount)
}

func Test_Demux_005(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Stop after 10 frames using io.EOF
	ctx := context.Background()
	frameCount := 0
	err = reader.Demux(ctx, nil, func(stream int, frame *ffmpeg.Frame) error {
		frameCount++
		if frameCount >= 10 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.Equal(10, frameCount)
	t.Logf("Stopped after %d frames", frameCount)
}

func Test_Demux_006(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Cancel context after short duration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	frameCount := 0
	err = reader.Demux(ctx, nil, func(stream int, frame *ffmpeg.Frame) error {
		frameCount++
		// Slow down to ensure timeout is hit
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	assert.Error(err)
	assert.Equal(context.DeadlineExceeded, err)
	t.Logf("Decoded %d frames before timeout", frameCount)
}

func Test_Demux_007(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Verify frame properties for video
	ctx := context.Background()
	foundValidPts := false
	foundValidDimensions := false
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.VIDEO {
			return par, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		if frame.Pts() >= 0 {
			foundValidPts = true
		}
		if frame.Width() > 0 && frame.Height() > 0 {
			foundValidDimensions = true
		}
		if foundValidPts && foundValidDimensions {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.True(foundValidPts, "Should find frames with valid PTS")
	assert.True(foundValidDimensions, "Should find frames with valid dimensions")
}

func Test_Demux_008(t *testing.T) {
	assert := assert.New(t)

	// Open audio file
	reader, err := ffmpeg.Open(TEST_WAV)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Verify frame properties for audio
	ctx := context.Background()
	foundSamples := false
	foundValidRate := false
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.AUDIO {
			return par, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		if frame.NumSamples() > 0 {
			foundSamples = true
		}
		if frame.SampleRate() > 0 {
			foundValidRate = true
		}
		if foundSamples && foundValidRate {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.True(foundSamples, "Should find frames with samples")
	assert.True(foundValidRate, "Should find frames with valid sample rate")
}

func Test_Demux_009(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Resample video to different size
	ctx := context.Background()
	targetWidth := 320
	targetHeight := 240
	frameCount := 0

	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.VIDEO {
			// Create new parameters with different size
			newPar, err := ffmpeg.NewVideoPar(
				par.PixelFormat().String(),
				"320x240",
				par.FrameRate(),
			)
			if err != nil {
				return nil, err
			}
			return newPar, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		assert.Equal(targetWidth, frame.Width())
		assert.Equal(targetHeight, frame.Height())
		frameCount++
		if frameCount >= 5 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.Equal(5, frameCount)
	t.Logf("Resampled %d frames to %dx%d", frameCount, targetWidth, targetHeight)
}

func Test_Demux_010(t *testing.T) {
	assert := assert.New(t)

	// Open audio file
	reader, err := ffmpeg.Open(TEST_WAV)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Resample audio to different sample rate
	ctx := context.Background()
	targetRate := 44100
	frameCount := 0

	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.AUDIO {
			// Get channel layout description
			ch := par.ChannelLayout()
			chStr, err := ff.AVUtil_channel_layout_describe(&ch)
			if err != nil {
				return nil, err
			}
			// Create new parameters with different sample rate
			newPar, err := ffmpeg.NewAudioPar(
				par.SampleFormat().String(),
				chStr,
				targetRate,
			)
			if err != nil {
				return nil, err
			}
			return newPar, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		assert.Equal(targetRate, frame.SampleRate())
		frameCount++
		if frameCount >= 5 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.Equal(5, frameCount)
	t.Logf("Resampled %d frames to %d Hz", frameCount, targetRate)
}

func Test_Demux_011(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)

	// Close reader
	err = reader.Close()
	assert.NoError(err)

	// Try to demux after close
	ctx := context.Background()
	err = reader.Demux(ctx, nil, func(stream int, frame *ffmpeg.Frame) error {
		assert.Fail("Should not be called")
		return nil
	})

	assert.Error(err)
	assert.Contains(err.Error(), "closed")
}

func Test_Demux_012(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Map function that returns error for specific stream
	ctx := context.Background()
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		return nil, nil // Ignore all streams
	}, func(stream int, frame *ffmpeg.Frame) error {
		assert.Fail("Should not be called when no streams mapped")
		return nil
	})

	// Should get error about no streams to decode
	assert.Error(err)
	assert.Contains(err.Error(), "no streams")
}

func Test_Demux_013(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Verify PTS timestamps are reasonable
	ctx := context.Background()
	lastPts := make(map[int]int64)
	nonDecreasingCount := 0
	totalFrames := 0

	err = reader.Demux(ctx, nil, func(stream int, frame *ffmpeg.Frame) error {
		pts := frame.Pts()
		if lastPts[stream] <= pts {
			nonDecreasingCount++
		}
		lastPts[stream] = pts
		totalFrames++
		if totalFrames >= 50 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.Greater(totalFrames, 0)
	t.Logf("Non-decreasing PTS: %d/%d frames", nonDecreasingCount, totalFrames)
}

func Test_Demux_014(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// First demux - count frames
	ctx := context.Background()
	firstCount := 0
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.VIDEO {
			return par, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		firstCount++
		return nil
	})
	assert.NoError(err)
	assert.Greater(firstCount, 0)

	// Seek back to start
	streams := reader.Streams(media.VIDEO)
	if len(streams) > 0 {
		err = reader.Seek(streams[0].Index(), 0)
		assert.NoError(err)
	}

	// Second demux should work and get similar count
	secondCount := 0
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if par.Type() == media.VIDEO {
			return par, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		secondCount++
		return nil
	})
	assert.NoError(err)
	assert.Greater(secondCount, 0)

	t.Logf("First demux: %d frames, second demux: %d frames", firstCount, secondCount)
	// Counts should be similar after seek
	assert.InDelta(firstCount, secondCount, float64(firstCount)*0.1) // Within 10%
}

func Test_Demux_015(t *testing.T) {
	assert := assert.New(t)

	// Open video file with both video and audio
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Track both video and audio frames
	ctx := context.Background()
	videoFrames := 0
	audioFrames := 0

	err = reader.Demux(ctx, nil, func(stream int, frame *ffmpeg.Frame) error {
		switch frame.Type() {
		case media.VIDEO:
			videoFrames++
		case media.AUDIO:
			audioFrames++
		}
		if videoFrames+audioFrames >= 50 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	t.Logf("Decoded %d video frames and %d audio frames", videoFrames, audioFrames)
	assert.Greater(videoFrames+audioFrames, 0)
}

func Test_Demux_016(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Map function that returns same parameters (no resampling needed)
	ctx := context.Background()
	frameCount := 0
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		// Return the same parameters - decoder should NOT create a resampler
		return par, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		frameCount++
		if frameCount >= 10 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.Equal(10, frameCount)
	t.Logf("Decoded %d frames without resampling (passthrough)", frameCount)
}

func Test_Demux_017(t *testing.T) {
	assert := assert.New(t)

	// Open video file with force flag
	reader, err := ffmpeg.Open(TEST_MP4, ffmpeg.OptForce())
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Map function that returns same parameters but force is set
	ctx := context.Background()
	frameCount := 0
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		// Return the same parameters - but with force=true, resampler should still be created
		return par, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		frameCount++
		if frameCount >= 10 {
			return io.EOF
		}
		return nil
	})

	assert.NoError(err)
	assert.Equal(10, frameCount)
	t.Logf("Decoded %d frames with force flag (resampler created even for matching formats)", frameCount)
}
