package ffmpeg_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg80"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

const (
	testOutputDir = "../../tmp"
	testInputMP4  = "../../etc/test/sample.mp4"
	testInputMP3  = "../../etc/test/sample.mp3"
	testInputWAV  = "../../etc/test/jfk.wav"
)

////////////////////////////////////////////////////////////////////////////////
// TEST REMUXING - Read packets and write directly (no transcoding)

func Test_remux_sync_mp4(t *testing.T) {
	assert := assert.New(t)

	// Create output directory
	if err := os.MkdirAll(testOutputDir, 0755); err != nil {
		t.Fatal(err)
	}

	inputFile := testInputMP4
	outputFile := filepath.Join(testOutputDir, "remux_sync.mp4")
	defer os.Remove(outputFile)

	// Open input file
	reader, err := ffmpeg.Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Create writer with streams matching input, building stream index map
	var writerOpts []ffmpeg.Opt
	writerOpts = append(writerOpts, ffmpeg.OptCopy()) // Enable copy mode
	streamMap := make(map[int]int)                    // input stream -> output stream
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &ffmpeg.Par{AVCodecParameters: *stream.CodecPar()}
		// Use 0 for auto-increment
		writerOpts = append(writerOpts, ffmpeg.OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	writer, err := ffmpeg.Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Decode and write packets synchronously with stream remapping
	ctx := context.Background()
	packetCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			(*ff.AVPacket)(pkt).SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets synchronously from %s to %s", packetCount, inputFile, outputFile)
}

func Test_remux_sync_mp3(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP3
	outputFile := filepath.Join(t.TempDir(), "remux_sync.mp3")

	reader, err := ffmpeg.Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	var writerOpts []ffmpeg.Opt
	writerOpts = append(writerOpts, ffmpeg.OptCopy())
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &ffmpeg.Par{AVCodecParameters: *stream.CodecPar()}
		// Use 0 for auto-increment
		writerOpts = append(writerOpts, ffmpeg.OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	writer, err := ffmpeg.Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	ctx := context.Background()
	packetCount := 0
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			(*ff.AVPacket)(pkt).SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets synchronously from %s to %s", packetCount, inputFile, outputFile)
}

func Test_remux_async_mp4(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP4
	outputFile := filepath.Join(t.TempDir(), "remux_async.mp4")

	reader, err := ffmpeg.Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	var writerOpts []ffmpeg.Opt
	writerOpts = append(writerOpts, ffmpeg.OptCopy())
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &ffmpeg.Par{AVCodecParameters: *stream.CodecPar()}
		// Use 0 for auto-increment
		writerOpts = append(writerOpts, ffmpeg.OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	writer, err := ffmpeg.Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	packets := make(chan *ffmpeg.Packet, 10)
	errChan := make(chan error, 2)

	go func() {
		errChan <- writer.WritePackets(packets)
	}()

	packetCount := 0
	go func() {
		defer close(packets)
		ctx := context.Background()
		err := reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
			if newStream, ok := streamMap[stream]; ok {
				(*ff.AVPacket)(pkt).SetStreamIndex(newStream)
			}
			// Ref packet before sending through channel (increment reference count)
			refPkt := ff.AVCodec_packet_alloc()
			if err := ff.AVCodec_packet_ref(refPkt, (*ff.AVPacket)(pkt)); err != nil {
				return err
			}
			packetCount++
			packets <- (*ffmpeg.Packet)(refPkt)
			return nil
		})
		errChan <- err
	}()

	err1 := <-errChan
	err2 := <-errChan

	assert.NoError(err1)
	assert.NoError(err2)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets asynchronously from %s to %s", packetCount, inputFile, outputFile)
}

func Test_remux_async_wav(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputWAV
	outputFile := filepath.Join(t.TempDir(), "remux_async.wav")

	reader, err := ffmpeg.Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	var writerOpts []ffmpeg.Opt
	writerOpts = append(writerOpts, ffmpeg.OptCopy())
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &ffmpeg.Par{AVCodecParameters: *stream.CodecPar()}
		// Use 0 for auto-increment
		writerOpts = append(writerOpts, ffmpeg.OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	writer, err := ffmpeg.Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	packets := make(chan *ffmpeg.Packet, 10)
	errChan := make(chan error, 2)

	go func() {
		errChan <- writer.WritePackets(packets)
	}()

	packetCount := 0
	go func() {
		defer close(packets)
		ctx := context.Background()
		err := reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
			if newStream, ok := streamMap[stream]; ok {
				(*ff.AVPacket)(pkt).SetStreamIndex(newStream)
			}
			// Ref packet before sending through channel (increment reference count)
			refPkt := ff.AVCodec_packet_alloc()
			if err := ff.AVCodec_packet_ref(refPkt, (*ff.AVPacket)(pkt)); err != nil {
				return err
			}
			packetCount++
			packets <- (*ffmpeg.Packet)(refPkt)
			return nil
		})
		errChan <- err
	}()

	err1 := <-errChan
	err2 := <-errChan

	assert.NoError(err1)
	assert.NoError(err2)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets asynchronously from %s to %s", packetCount, inputFile, outputFile)
}

////////////////////////////////////////////////////////////////////////////////
// TEST SYNCHRONOUS ENCODING - Direct writer.Write()

func Test_encoder_sync_audio(t *testing.T) {
	t.Skip("Skipping - needs frame generation implementation")
}

func Test_encoder_sync_video(t *testing.T) {
	t.Skip("Skipping - needs frame generation implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST ASYNCHRONOUS ENCODING - Channel-based writer.WritePackets()

func Test_encoder_async_audio(t *testing.T) {
	t.Skip("Skipping - needs frame generation implementation")
}

func Test_encoder_async_video(t *testing.T) {
	t.Skip("Skipping - needs frame generation implementation")
}

func Test_encoder_async_multi_stream(t *testing.T) {
	t.Skip("Skipping - needs frame generation implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST ERROR CASES

func Test_encoder_write_nil_packet(t *testing.T) {
	t.Skip("Skipping - needs valid writer instance")
}

func Test_encoder_write_packets_nil_channel(t *testing.T) {
	t.Skip("Skipping - needs valid writer instance")
}

func Test_encoder_write_packets_empty_channel(t *testing.T) {
	t.Skip("Skipping - needs valid writer instance")
}
