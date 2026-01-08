package ffmpeg_test

import (
	"context"
	"io"
	"testing"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

const (
	TEST_SRT = "../../etc/test/sample.srt"
)

func Test_FrameBuffer_001(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Get video stream for buffer creation
	var videoStream *schema.Stream
	for _, stream := range reader.Streams(media.VIDEO) {
		videoStream = stream
		break
	}
	assert.NotNil(videoStream, "no video stream found")

	// Create frame buffer with 2 second duration, common timebase 1/90000
	timebase := ff.AVUtil_rational_d2q(1.0/90000.0, 0)
	buffer, err := ffmpeg.NewFrameBuffer(timebase, 2*time.Second, videoStream)
	assert.NoError(err)
	assert.NotNil(buffer)

	// Decode frames and push into buffer
	ctx := context.Background()
	frameCount := 0
	pushCount := 0

	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		// Only decode video stream
		if stream == videoStream.Index() {
			return par, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		frameCount++

		// Create schema.Frame wrapper
		schemaFrame := schema.NewFrame((*ff.AVFrame)(frame), stream)
		assert.NotNil(schemaFrame)

		// Try to push into buffer
		err := buffer.Push(schemaFrame)
		if err == nil {
			pushCount++
		} else if err == ffmpeg.ErrBufferFull {
			// Buffer full - drain one frame and retry
			oldest, _ := buffer.Next(-1)
			if oldest != nil {
				oldest.Unref()
			}
			err = buffer.Push(schemaFrame)
			if err == nil {
				pushCount++
			}
		}

		// Stop after pushing 20 frames
		if pushCount >= 20 {
			return io.EOF
		}
		return err
	}, nil)

	assert.True(err == nil || err == io.EOF)
	assert.Greater(frameCount, 0)
	assert.Greater(pushCount, 0)

	// Get buffer stats
	stats := buffer.Stats()
	assert.Greater(stats.TotalFrames, 0)
	assert.GreaterOrEqual(stats.OldestPTS, int64(0))
	assert.Greater(stats.NewestPTS, int64(0))
	assert.Greater(stats.Duration, int64(0))

	t.Logf("Decoded %d frames, pushed %d into buffer", frameCount, pushCount)
	t.Logf("Buffer stats: frames=%d, duration=%d, oldest=%d, newest=%d",
		stats.TotalFrames, stats.Duration, stats.OldestPTS, stats.NewestPTS)

	// Drain buffer
	drainCount := 0
	for {
		frame, err := buffer.Next(-1)
		if err == io.EOF {
			break
		}
		assert.NoError(err)
		if frame == nil {
			break
		}
		frame.Unref()
		drainCount++
	}

	t.Logf("Drained %d frames from buffer", drainCount)
	assert.Equal(pushCount, drainCount)
}

func Test_FrameBuffer_002(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	assert.NotNil(reader)
	defer reader.Close()

	// Get video and audio streams
	videoStreams := reader.Streams(media.VIDEO)
	audioStreams := reader.Streams(media.AUDIO)

	var videoStream, audioStream *schema.Stream
	if len(videoStreams) > 0 {
		videoStream = videoStreams[0]
	}
	if len(audioStreams) > 0 {
		audioStream = audioStreams[0]
	}
	assert.NotNil(videoStream, "no video stream found")

	// Create frame buffer with both streams if audio exists
	timebase := ff.AVUtil_rational_d2q(1.0/90000.0, 0)
	var streams []*schema.Stream
	if audioStream != nil {
		streams = []*schema.Stream{videoStream, audioStream}
	} else {
		streams = []*schema.Stream{videoStream}
	}

	buffer, err := ffmpeg.NewFrameBuffer(timebase, 1*time.Second, streams...)
	assert.NoError(err)
	assert.NotNil(buffer)

	// Decode frames from both streams
	ctx := context.Background()
	framesByStream := make(map[int]int)
	totalPushed := 0

	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		// Decode video and audio
		if stream == videoStream.Index() || (audioStream != nil && stream == audioStream.Index()) {
			return par, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		framesByStream[stream]++

		// Create schema.Frame wrapper
		schemaFrame := schema.NewFrame((*ff.AVFrame)(frame), stream)

		// Push into buffer
		err := buffer.Push(schemaFrame)
		if err == nil {
			totalPushed++
		} else if err != ffmpeg.ErrBufferFull {
			return err
		}

		// Stop after 50 total frames
		if totalPushed >= 50 {
			return io.EOF
		}
		return nil
	}, nil)

	assert.True(err == nil || err == io.EOF)
	assert.Greater(len(framesByStream), 0)

	t.Logf("Decoded frames by stream: %v", framesByStream)
	t.Logf("Total pushed: %d", totalPushed)

	// Verify buffer can retrieve frames in PTS order
	var lastPTS int64 = -1
	retrieveCount := 0
	for {
		frame, err := buffer.Next(lastPTS)
		if err == io.EOF {
			break
		}
		assert.NoError(err)
		if frame == nil {
			break
		}

		// Verify PTS is increasing
		assert.Greater(frame.Pts, lastPTS, "frames should be in PTS order")
		lastPTS = frame.Pts

		frame.Unref()
		retrieveCount++
	}

	t.Logf("Retrieved %d frames in PTS order", retrieveCount)
}

func Test_FrameBuffer_Flush(t *testing.T) {
	assert := assert.New(t)

	// Open video file
	reader, err := ffmpeg.Open(TEST_MP4)
	assert.NoError(err)
	defer reader.Close()

	// Get video stream
	videoStreams := reader.Streams(media.VIDEO)
	assert.Greater(len(videoStreams), 0)
	videoStream := videoStreams[0]

	// Create buffer
	timebase := ff.AVUtil_rational_d2q(1.0/90000.0, 0)
	buffer, err := ffmpeg.NewFrameBuffer(timebase, 1*time.Second, videoStream)
	assert.NoError(err)

	// Decode and push some frames
	ctx := context.Background()
	pushCount := 0

	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == videoStream.Index() {
			return par, nil
		}
		return nil, nil
	}, func(stream int, frame *ffmpeg.Frame) error {
		schemaFrame := schema.NewFrame((*ff.AVFrame)(frame), stream)
		err := buffer.Push(schemaFrame)
		if err == nil {
			pushCount++
		}
		if pushCount >= 10 {
			return io.EOF
		}
		return err
	}, nil)

	assert.True(err == nil || err == io.EOF)
	assert.Equal(10, pushCount)

	// Verify buffer has frames
	stats := buffer.Stats()
	assert.Greater(stats.TotalFrames, 0)
	t.Logf("Buffer has %d frames before flush", stats.TotalFrames)

	// Flush buffer
	buffer.Flush()

	// Verify buffer is empty
	stats = buffer.Stats()
	assert.Equal(0, stats.TotalFrames)
	assert.Equal(int64(-1), stats.OldestPTS)
	assert.Equal(int64(-1), stats.NewestPTS)

	// Next should return nil after flush
	frame, err := buffer.Next(-1)
	assert.NoError(err)
	assert.Nil(frame)

	t.Log("Buffer successfully flushed")
}

func Test_FrameBuffer_Subtitles(t *testing.T) {
	assert := assert.New(t)

	// Open subtitle file
	reader, err := ffmpeg.Open(TEST_SRT)
	assert.NoError(err)
	defer reader.Close()

	// Get subtitle stream
	subtitleStreams := reader.Streams(media.SUBTITLE)
	assert.Greater(len(subtitleStreams), 0, "no subtitle stream found")
	subtitleStream := subtitleStreams[0]

	t.Logf("Subtitle stream codec: %s", subtitleStream.CodecPar().CodecID().Name())
	t.Logf("Subtitle stream timebase: %d/%d", subtitleStream.TimeBase().Num(), subtitleStream.TimeBase().Den())

	// Create buffer for subtitles (1 minute duration)
	timebase := ff.AVUtil_rational_d2q(1.0/1000.0, 0) // millisecond timebase
	buffer, err := ffmpeg.NewFrameBuffer(timebase, 1*time.Minute, subtitleStream)
	assert.NoError(err)

	// Decode and push subtitles
	ctx := context.Background()
	subtitleCount := 0
	pushCount := 0
	packetCount := 0

	// First, read packets to see what we're getting
	t.Log("Reading packets from subtitle file...")
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		packetCount++
		// Show packet data to understand format
		data := pkt.Bytes()
		dataPreview := string(data)
		if len(dataPreview) > 50 {
			dataPreview = dataPreview[:50] + "..."
		}
		t.Logf("Packet %d: stream=%d, pts=%d, dts=%d, size=%d, data=%q",
			packetCount, stream, pkt.Pts(), pkt.Dts(), pkt.Size(), dataPreview)
		if packetCount >= 10 {
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		t.Logf("Error reading packets: %v", err)
	}

	// Reset reader by closing and reopening
	reader.Close()
	reader, err = ffmpeg.Open(TEST_SRT)
	assert.NoError(err)
	defer reader.Close()

	// Now try to decode subtitles
	t.Log("Attempting to decode subtitles...")
	err = reader.Demux(ctx, func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == subtitleStream.Index() {
			t.Logf("Mapping subtitle stream %d with codec %s", stream, par.CodecID().Name())
			return par, nil
		}
		return nil, nil
	}, nil, func(stream int, subtitle *ff.AVSubtitle) error {
		subtitleCount++
		t.Logf("Decoded subtitle %d from stream %d", subtitleCount, stream)

		// Create schema.Frame wrapper for subtitle
		schemaFrame := schema.NewSubtitle(subtitle, stream)
		assert.NotNil(schemaFrame)
		assert.NotNil(schemaFrame.AVSubtitle)
		assert.Nil(schemaFrame.AVFrame)

		// Create multiple refs to test reference counting
		ref1 := schemaFrame.Ref()
		assert.NotNil(ref1)
		assert.Equal(schemaFrame, ref1) // Should be same pointer for subtitles

		ref2 := schemaFrame.Ref()
		assert.NotNil(ref2)

		// Push original into buffer (takes another ref internally via Ref())
		err := buffer.Push(schemaFrame)
		assert.NoError(err)
		pushCount++

		// Unref the extra refs we took
		ref1.Unref()
		ref2.Unref()

		// Original will be unreffed when decode callback returns
		return nil
	})

	// Subtitle codec might not be available in all FFmpeg builds
	if err != nil {
		t.Logf("Demux error: %v", err)
		if subtitleCount == 0 {
			t.Skip("Subtitle decoding not supported (codec may not be compiled in FFmpeg build)")
			return
		}
	}
	assert.NoError(err)
	assert.Greater(subtitleCount, 0)
	assert.Equal(subtitleCount, pushCount)

	t.Logf("Decoded %d subtitles, pushed %d into buffer", subtitleCount, pushCount)

	// Get buffer stats
	stats := buffer.Stats()
	assert.Equal(pushCount, stats.TotalFrames)
	assert.GreaterOrEqual(stats.OldestPTS, int64(0))
	assert.Greater(stats.NewestPTS, int64(0))

	t.Logf("Buffer stats: frames=%d, duration=%d, oldest=%d, newest=%d",
		stats.TotalFrames, stats.Duration, stats.OldestPTS, stats.NewestPTS)

	// Drain buffer and verify frames come out in order
	drainCount := 0
	var lastPTS int64 = -1
	for {
		frame, err := buffer.Next(lastPTS)
		if err == io.EOF {
			break
		}
		assert.NoError(err)
		if frame == nil {
			break
		}

		// Verify it's a subtitle frame
		assert.NotNil(frame.AVSubtitle)
		assert.Nil(frame.AVFrame)

		// Verify PTS ordering
		assert.Greater(frame.Pts, lastPTS, "subtitles should be in PTS order")
		lastPTS = frame.Pts

		// Unref will decrement refcount, free when it hits 0
		frame.Unref()
		drainCount++
	}

	t.Logf("Drained %d subtitle frames from buffer", drainCount)
	assert.Equal(pushCount, drainCount)

	// Verify buffer is empty
	stats = buffer.Stats()
	assert.Equal(0, stats.TotalFrames)
}
