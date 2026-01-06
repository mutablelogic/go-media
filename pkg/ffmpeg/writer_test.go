package ffmpeg

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

const (
	testInputMP3    = "../../etc/test/sample.mp3"
	testInputMP4    = "../../etc/test/sample.mp4"
	testArtworkFile = "../../etc/test/sample.jpg"
)

func fillVideoBlackYUV420P(frame *Frame) {
	if frame == nil {
		return
	}
	// Y plane: 0, U/V planes: 128 for neutral chroma.
	y := frame.Bytes(0)
	for i := range y {
		y[i] = 0
	}
	u := frame.Bytes(1)
	for i := range u {
		u[i] = 128
	}
	v := frame.Bytes(2)
	for i := range v {
		v[i] = 128
	}
}

func fillAudioSilenceFLTP(frame *Frame) {
	if frame == nil {
		return
	}
	channels := (*ff.AVFrame)(frame).NumChannels()
	for ch := 0; ch < channels; ch++ {
		plane := frame.Float32(ch)
		for i := range plane {
			plane[i] = 0
		}
	}
}

func assertStreamOrderVideoAudio(t *testing.T, w *Writer) {
	assert := assert.New(t)
	if !assert.NotNil(w) {
		return
	}
	enc0 := w.Stream(0)
	enc1 := w.Stream(1)
	if !assert.NotNil(enc0, "expected encoder for stream 0") {
		return
	}
	if !assert.NotNil(enc1, "expected encoder for stream 1") {
		return
	}
	par0 := enc0.Par()
	par1 := enc1.Par()
	if !assert.NotNil(par0) || !assert.NotNil(par1) {
		return
	}
	assert.Equal(ff.AVMEDIA_TYPE_VIDEO, par0.CodecType(), "stream 0 should be video")
	assert.Equal(ff.AVMEDIA_TYPE_AUDIO, par1.CodecType(), "stream 1 should be audio")
}

////////////////////////////////////////////////////////////////////////////////
// TEST CREATE FILE

func Test_writer_create_mp4(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

func Test_writer_create_mp3(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

func Test_writer_create_wav(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST CREATE WITH METADATA

func Test_writer_create_with_metadata(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

func Test_writer_create_with_artwork(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST CREATE WITH MULTIPLE STREAMS

func Test_writer_create_multiple_streams(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST NEW WRITER WITH IO.WRITER

func Test_writer_new_buffer(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

func Test_writer_new_file(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST ERROR CASES

func Test_writer_create_no_format(t *testing.T) {
	assert := assert.New(t)

	// Try to create writer with no format hint
	w, err := Create("", OptStream(0, nil))
	assert.Error(err)
	assert.Nil(w)
}

func Test_writer_create_no_streams(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(t.TempDir(), "test_empty.mp4")

	// Try to create writer without streams
	w, err := Create(testFile)
	assert.Error(err)
	assert.Nil(w)
	assert.Contains(err.Error(), "no streams")
}

func Test_writer_create_invalid_codec(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(t.TempDir(), "test_invalid.mp4")

	// Try to create with invalid codec
	audioPar, err := NewAudioPar("invalid_codec", "stereo", 44100)
	if err != nil {
		t.Skip("Invalid codec properly rejected:", err)
	}

	w, err := Create(testFile, OptStream(0, audioPar))
	assert.Error(err)
	assert.Nil(w)
}

func Test_writer_create_duplicate_stream(t *testing.T) {
	t.Skip("Skipping - duplicate stream validation not implemented")
}

////////////////////////////////////////////////////////////////////////////////
// TEST STREAM ACCESS

func Test_writer_stream_access(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

func Test_writer_close_twice(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST OUTPUT FORMAT GUESSING

func Test_writer_format_guess_mp4(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

func Test_writer_format_guess_mkv(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

func Test_writer_format_explicit(t *testing.T) {
	t.Skip("Skipping - needs frame encoding implementation")
}

////////////////////////////////////////////////////////////////////////////////
// TEST METADATA AND ARTWORK COPYING

// Test copying metadata from original file and adding new metadata
func Test_writer_copy_and_add_metadata_mp3(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP3
	outputFile := filepath.Join(t.TempDir(), "copy_metadata.mp3")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Get existing metadata from input
	existingMeta := reader.Metadata()
	t.Logf("Found %d metadata entries in input file", len(existingMeta))

	// Create writer with copy mode
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())

	// Copy existing metadata
	for _, meta := range existingMeta {
		if meta.Key() != MetaArtwork { // Skip artwork for now
			writerOpts = append(writerOpts, OptMetadata(meta))
		}
	}

	// Add new metadata
	writerOpts = append(writerOpts,
		OptMetadata(NewMetadata("custom_tag", "test_value")),
		OptMetadata(NewMetadata("test_artist", "Test Artist")),
		OptMetadata(NewMetadata("test_album", "Test Album")),
	)

	// Add streams from input
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	// Create writer
	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Copy packets
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets with copied and new metadata from %s to %s", packetCount, inputFile, outputFile)

	// Close writer before verification
	if err := writer.Close(); !assert.NoError(err) {
		t.Logf("Error closing writer: %v", err)
		t.FailNow()
	}

	// Verify output file exists
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))

	// Read back and verify metadata
	verifyReader, err := Open(outputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer verifyReader.Close()

	outputMeta := verifyReader.Metadata()
	t.Logf("Output file has %d metadata entries", len(outputMeta))
	assert.GreaterOrEqual(len(outputMeta), len(existingMeta)+3, "Should have original + new metadata")
}

// Test copying artwork from original and replacing with new artwork
func Test_writer_copy_and_replace_artwork_mp3(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP3
	outputFile := filepath.Join(t.TempDir(), "replace_artwork.mp3")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Get existing metadata (including artwork)
	existingMeta := reader.Metadata()
	hasOriginalArtwork := false
	for _, meta := range existingMeta {
		if meta.Key() == MetaArtwork {
			hasOriginalArtwork = true
			t.Logf("Original file has artwork (%d bytes)", len(meta.Bytes()))
		}
	}

	// Read new artwork
	newArtworkData, err := os.ReadFile(testArtworkFile)
	if !assert.NoError(err) {
		t.Skip("New artwork file not available:", err)
	}
	t.Logf("New artwork is %d bytes", len(newArtworkData))

	// Create writer with copy mode
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())

	// Copy metadata (excluding artwork)
	for _, meta := range existingMeta {
		if meta.Key() != MetaArtwork {
			writerOpts = append(writerOpts, OptMetadata(meta))
		}
	}

	// Add new artwork (replaces original)
	writerOpts = append(writerOpts, OptMetadata(NewMetadata(MetaArtwork, newArtworkData)))

	// Add streams from input
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	// Create writer
	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Copy packets
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets with replaced artwork from %s to %s", packetCount, inputFile, outputFile)

	// Close writer before verification
	assert.NoError(writer.Close())

	// Verify output file exists
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))

	// Read back and verify artwork was replaced
	verifyReader, err := Open(outputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer verifyReader.Close()

	outputMeta := verifyReader.Metadata(MetaArtwork)
	if hasOriginalArtwork {
		assert.Len(outputMeta, 1, "Should have one artwork entry")
		if len(outputMeta) > 0 && outputMeta[0].Key() == MetaArtwork {
			artworkBytes := outputMeta[0].Bytes()
			t.Logf("Output artwork is %d bytes", len(artworkBytes))
			// New artwork should be different size than original
			assert.Equal(len(newArtworkData), len(artworkBytes), "Artwork should be replaced")
		}
	}
}

// Test copying all metadata and artwork, plus adding new metadata
func Test_writer_copy_all_add_metadata_mp4(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP4
	outputFile := filepath.Join(t.TempDir(), "copy_all_metadata.mp4")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Get ALL existing metadata from input (including artwork)
	existingMeta := reader.Metadata()
	t.Logf("Found %d metadata entries in input file", len(existingMeta))

	// Create writer with copy mode
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())

	// Copy ALL existing metadata (including artwork)
	for _, meta := range existingMeta {
		writerOpts = append(writerOpts, OptMetadata(meta))
	}

	// Add new metadata entries
	writerOpts = append(writerOpts,
		OptMetadata(NewMetadata("encoder", "go-media test suite")),
		OptMetadata(NewMetadata("comment", "Test file with copied and new metadata")),
	)

	// Add streams from input
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	// Create writer
	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Copy packets
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets with all metadata copied plus new entries from %s to %s", packetCount, inputFile, outputFile)

	// Close writer before verification
	assert.NoError(writer.Close())

	// Verify output file exists
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))

	// Read back and verify metadata
	verifyReader, err := Open(outputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer verifyReader.Close()

	outputMeta := verifyReader.Metadata()
	t.Logf("Output file has %d metadata entries", len(outputMeta))
	assert.GreaterOrEqual(len(outputMeta), len(existingMeta), "Should have at least original metadata")
}

// Test adding artwork to file that originally had none
func Test_writer_add_new_artwork_mp4(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP4
	outputFile := filepath.Join(t.TempDir(), "add_artwork.mp4")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Check if input has artwork
	existingMeta := reader.Metadata(MetaArtwork)
	if len(existingMeta) > 0 {
		t.Log("Input file already has artwork, will replace it")
	} else {
		t.Log("Input file has no artwork, will add new")
	}

	// Read new artwork
	artworkData, err := os.ReadFile(testArtworkFile)
	if !assert.NoError(err) {
		t.Skip("Artwork file not available:", err)
	}

	// Create writer with copy mode
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())

	// Copy metadata (excluding any existing artwork)
	for _, meta := range reader.Metadata() {
		if meta.Key() != MetaArtwork {
			writerOpts = append(writerOpts, OptMetadata(meta))
		}
	}

	// Add new artwork
	writerOpts = append(writerOpts, OptMetadata(NewMetadata(MetaArtwork, artworkData)))

	// Add streams from input
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	// Create writer
	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Copy packets
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets with new artwork from %s to %s", packetCount, inputFile, outputFile)

	// Close writer before verification
	assert.NoError(writer.Close())

	// Verify output file exists and has artwork
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))

	// Verify artwork was added
	verifyReader, err := Open(outputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer verifyReader.Close()

	outputArtwork := verifyReader.Metadata(MetaArtwork)
	assert.Len(outputArtwork, 1, "Should have artwork")
	if len(outputArtwork) > 0 {
		artworkBytes := outputArtwork[0].Bytes()
		t.Logf("Output artwork is %d bytes", len(artworkBytes))
		assert.Greater(len(artworkBytes), 0)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST ARTWORK REMUX

func Test_artwork_remux_mp3(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP3
	outputFile := filepath.Join(t.TempDir(), "artwork.mp3")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Read artwork
	artworkData, err := os.ReadFile(testArtworkFile)
	if !assert.NoError(err) {
		t.Skip("Artwork file not available:", err)
	}

	// Create writer with artwork metadata
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())
	writerOpts = append(writerOpts, OptMetadata(NewMetadata(MetaArtwork, artworkData)))

	// Add streams from input
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	// Create writer
	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Copy packets (artwork will be written automatically on first packet)
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets with artwork from %s to %s", packetCount, inputFile, outputFile)

	// Verify output file exists and has artwork
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
}

func Test_artwork_remux_mp4(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP4
	outputFile := filepath.Join(t.TempDir(), "artwork.mp4")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Read artwork
	artworkData, err := os.ReadFile(testArtworkFile)
	if !assert.NoError(err) {
		t.Skip("Artwork file not available:", err)
	}

	// Create writer with artwork metadata
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())
	writerOpts = append(writerOpts, OptMetadata(NewMetadata(MetaArtwork, artworkData)))

	// Add streams from input
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	// Create writer
	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Copy packets (artwork will be written automatically on first packet)
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets with artwork from %s to %s", packetCount, inputFile, outputFile)

	// Verify output file exists
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
}

////////////////////////////////////////////////////////////////////////////////
// TEST ARTWORK WITHOUT DATA

func Test_artwork_no_data(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP3
	outputFile := filepath.Join(t.TempDir(), "no_artwork.mp3")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Create writer without artwork
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())

	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Copy packets normally (no artwork, automatic check will skip it)
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME ENCODING

func Test_encode_frame_single(t *testing.T) {
	t.Skip("Skipping - needs frame generation implementation")
}

func Test_encode_frames_channel(t *testing.T) {
	t.Skip("Skipping - needs frame generation implementation")
}

// Test encoding black video frames to MP4
func Test_encode_black_video_mp4(t *testing.T) {
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "black_video.mp4")

	// Create video parameters (720p, 30fps, H.264)
	videoPar, err := NewVideoPar("yuv420p", "1280x720", 30.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with video stream
	writer, err := Create(outputFile, OptStream(0, videoPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Create a frame with video parameters
	frame, err := NewFrame(videoPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Allocate buffers (will be zeroed/black)
	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	// Encode 30 black frames (1 second at 30fps)
	for i := 0; i < 30; i++ {
		frame.SetPts(int64(i))
		if err := writer.EncodeFrame(0, frame); !assert.NoError(err) {
			t.FailNow()
		}
	}

	// Flush encoder
	if err := writer.EncodeFrame(0, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
	t.Logf("Created black video: %s (%d bytes)", outputFile, info.Size())
}

// Test encoding silent audio frames to M4A (AAC)
func Test_encode_silent_audio_mp3(t *testing.T) {
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "silent_audio.m4a")

	// Create audio parameters (stereo, 44.1kHz, AAC uses fltp = floating point planar)
	audioPar, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with audio stream
	writer, err := Create(outputFile, OptStream(0, audioPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Create a frame with audio parameters
	frame, err := NewFrame(audioPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Set number of samples per frame (1024 is common for AAC)
	(*ff.AVFrame)(frame).SetNumSamples(1024)

	// Allocate buffers (will be zeroed/silent)
	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	// Encode 100 silent frames (~2.3 seconds at 44.1kHz)
	for i := 0; i < 100; i++ {
		frame.SetPts(int64(i * 1024))
		if err := writer.EncodeFrame(0, frame); !assert.NoError(err) {
			t.FailNow()
		}
	}

	// Flush encoder
	if err := writer.EncodeFrame(0, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
	t.Logf("Created silent audio (AAC): %s (%d bytes)", outputFile, info.Size())
}

// Test encoding frames via channel (asynchronous)
func Test_encode_frames_async_mp4(t *testing.T) {
	// Fixed: encoder now leaves packet ownership to muxer
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "async_video.mp4")

	// Create video parameters (480p, 25fps, H.264)
	videoPar, err := NewVideoPar("yuv420p", "640x480", 25.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with video stream
	writer, err := Create(outputFile, OptStream(0, videoPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Create channel for frames
	frames := make(chan *Frame) // Unbuffered to force sync

	// Generate and send frames in a goroutine
	go func() {
		defer close(frames)

		for i := 0; i < 50; i++ { // 2 seconds at 25fps
			frame, err := NewFrame(videoPar)
			if err != nil {
				t.Logf("Failed to create frame: %v", err)
				return
			}

			if err := frame.AllocateBuffers(); err != nil {
				frame.Close()
				t.Logf("Failed to allocate buffers: %v", err)
				return
			}
			fillVideoBlackYUV420P(frame)

			frame.SetPts(int64(i))
			// Set duration to 1 (in timebase units, so 1/25s)
			(*ff.AVFrame)(frame).SetDuration(1)
			frames <- frame
		}
	}()

	// Run encoding on the main thread
	encodeErr := writer.EncodeFrames(0, frames)
	assert.NoError(encodeErr)

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
	t.Logf("Created async video: %s (%d bytes)", outputFile, info.Size())
}

// Test encoding multiple streams (video + audio) to MP4
func Test_encode_multiple_streams_mp4(t *testing.T) {
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "multi_stream.mp4")

	// Create video parameters (720p, 30fps, H.264)
	videoPar, err := NewVideoPar("yuv420p", "1280x720", 30.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create audio parameters (stereo, 44.1kHz, AAC)
	audioPar, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with both video and audio streams
	// Use stream=0 for auto-assignment (will be stream indices 0 and 1)
	writer, err := Create(outputFile, OptStream(0, videoPar), OptStream(0, audioPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()
	assertStreamOrderVideoAudio(t, writer)

	// Create video frame
	videoFrame, err := NewFrame(videoPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer videoFrame.Close()

	if err := videoFrame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}
	fillVideoBlackYUV420P(videoFrame)

	// Create audio frame
	audioFrame, err := NewFrame(audioPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer audioFrame.Close()

	(*ff.AVFrame)(audioFrame).SetNumSamples(1024)
	if err := audioFrame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}
	fillAudioSilenceFLTP(audioFrame)

	// Encode 60 video frames (2 seconds at 30fps)
	// and corresponding audio frames (~2 seconds at 44.1kHz with 1024 samples per frame)
	numVideoFrames := 60
	samplesPerAudioFrame := 1024
	totalAudioSamples := int64(numVideoFrames) * 44100 / 30 // Audio samples for 2 seconds
	numAudioFrames := int(totalAudioSamples / int64(samplesPerAudioFrame))

	// Encode video frames
	for i := 0; i < numVideoFrames; i++ {
		videoFrame.SetPts(int64(i))
		if err := writer.EncodeFrame(0, videoFrame); !assert.NoError(err) {
			t.FailNow()
		}
	}

	// Flush video encoder
	if err := writer.EncodeFrame(0, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Encode audio frames
	for i := 0; i < numAudioFrames; i++ {
		audioFrame.SetPts(int64(i * samplesPerAudioFrame))
		if err := writer.EncodeFrame(1, audioFrame); !assert.NoError(err) {
			t.FailNow()
		}
	}

	// Flush audio encoder
	if err := writer.EncodeFrame(1, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
	t.Logf("Created multi-stream video: %s (%d bytes, %d video frames, %d audio frames)",
		outputFile, info.Size(), numVideoFrames, numAudioFrames)
}

// Test encoding multiple streams asynchronously (video + audio via channels)
func Test_encode_multiple_streams_async_mp4(t *testing.T) {
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "multi_stream_async.mp4")

	// Create video parameters (720p, 30fps, H.264)
	videoPar, err := NewVideoPar("yuv420p", "1280x720", 30.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create audio parameters (stereo, 44.1kHz, AAC)
	audioPar, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with both video and audio streams
	writer, err := Create(outputFile, OptStream(0, videoPar), OptStream(0, audioPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()
	assertStreamOrderVideoAudio(t, writer)

	// Create channels for frames
	videoFrames := make(chan *Frame, 10)
	audioFrames := make(chan *Frame, 10)

	// Start encoding goroutines
	var videoErr, audioErr error
	videoDone := make(chan struct{})
	audioDone := make(chan struct{})

	go func() {
		defer close(videoDone)
		videoErr = writer.EncodeFrames(0, videoFrames)
	}()

	go func() {
		defer close(audioDone)
		audioErr = writer.EncodeFrames(1, audioFrames)
	}()

	// Generate video frames in goroutine
	numVideoFrames := 60
	go func() {
		defer close(videoFrames)
		for i := 0; i < numVideoFrames; i++ {
			frame, err := NewFrame(videoPar)
			if err != nil {
				t.Logf("Failed to create video frame: %v", err)
				return
			}
			if err := frame.AllocateBuffers(); err != nil {
				frame.Close()
				t.Logf("Failed to allocate video buffers: %v", err)
				return
			}
			fillVideoBlackYUV420P(frame)
			frame.SetPts(int64(i))
			videoFrames <- frame
		}
	}()

	// Generate audio frames in goroutine
	samplesPerAudioFrame := 1024
	totalAudioSamples := int64(numVideoFrames) * 44100 / 30
	// For AAC, only send complete frames (ignore remainder)
	numAudioFrames := int(totalAudioSamples / int64(samplesPerAudioFrame))

	go func() {
		defer close(audioFrames)
		for i := 0; i < numAudioFrames; i++ {
			frame, err := NewFrame(audioPar)
			if err != nil {
				t.Logf("Failed to create audio frame: %v", err)
				return
			}

			(*ff.AVFrame)(frame).SetNumSamples(samplesPerAudioFrame)
			if err := frame.AllocateBuffers(); err != nil {
				frame.Close()
				t.Logf("Failed to allocate audio buffers: %v", err)
				return
			}
			fillAudioSilenceFLTP(frame)
			frame.SetPts(int64(i * samplesPerAudioFrame))
			audioFrames <- frame
		}
	}()

	// Wait for encoding to complete
	<-videoDone
	<-audioDone
	assert.NoError(videoErr)
	assert.NoError(audioErr)

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
	t.Logf("Created async multi-stream video: %s (%d bytes, %d video frames, %d audio frames)",
		outputFile, info.Size(), numVideoFrames, numAudioFrames)
}

// Test encoding multiple streams with mixed sync/async (video async, audio sync)
func Test_encode_multiple_streams_mixed_mp4(t *testing.T) {
	// Fixed: encoder now leaves packet ownership to muxer
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "multi_stream_mixed.mp4")

	// Create video parameters (640x480, 25fps, H.264)
	videoPar, err := NewVideoPar("yuv420p", "640x480", 25.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create audio parameters (stereo, 44.1kHz, AAC)
	audioPar, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with both video and audio streams
	writer, err := Create(outputFile, OptStream(0, videoPar), OptStream(0, audioPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()
	assertStreamOrderVideoAudio(t, writer)

	// Create channel for video frames (async)
	videoFrames := make(chan *Frame, 10)

	// Start video encoding goroutine
	var videoErr error
	videoDone := make(chan struct{})

	go func() {
		defer close(videoDone)
		videoErr = writer.EncodeFrames(0, videoFrames)
	}()

	// Generate video frames in goroutine
	numVideoFrames := 50
	go func() {
		defer close(videoFrames)
		for i := 0; i < numVideoFrames; i++ {
			frame, err := NewFrame(videoPar)
			if err != nil {
				t.Logf("Failed to create video frame: %v", err)
				return
			}
			if err := frame.AllocateBuffers(); err != nil {
				frame.Close()
				t.Logf("Failed to allocate video buffers: %v", err)
				return
			}
			fillVideoBlackYUV420P(frame)
			frame.SetPts(int64(i))
			videoFrames <- frame
		}
	}()

	// Encode audio frames synchronously
	audioFrame, err := NewFrame(audioPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer audioFrame.Close()

	samplesPerAudioFrame := 1024
	(*ff.AVFrame)(audioFrame).SetNumSamples(samplesPerAudioFrame)
	if err := audioFrame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}
	fillAudioSilenceFLTP(audioFrame)

	totalAudioSamples := int64(numVideoFrames) * 44100 / 25
	numAudioFrames := int(totalAudioSamples / int64(samplesPerAudioFrame))

	for i := 0; i < numAudioFrames; i++ {
		audioFrame.SetPts(int64(i * samplesPerAudioFrame))
		if err := writer.EncodeFrame(1, audioFrame); !assert.NoError(err) {
			t.FailNow()
		}
	}

	// Flush audio encoder
	if err := writer.EncodeFrame(1, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Wait for video encoding to complete
	<-videoDone
	assert.NoError(videoErr)

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
	t.Logf("Created mixed sync/async multi-stream video: %s (%d bytes, %d video frames, %d audio frames)",
		outputFile, info.Size(), numVideoFrames, numAudioFrames)
}

////////////////////////////////////////////////////////////////////////////////
// TEST MULTIPLE ARTWORKS

func Test_artwork_multiple_mp3(t *testing.T) {
	assert := assert.New(t)

	inputFile := testInputMP3
	outputFile := filepath.Join(t.TempDir(), "multiple_artwork.mp3")

	// Read input file
	reader, err := Open(inputFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer reader.Close()

	// Read artwork
	artworkData, err := os.ReadFile(testArtworkFile)
	if !assert.NoError(err) {
		t.Skip("Artwork file not available:", err)
	}

	// Create writer with MULTIPLE artwork entries
	var writerOpts []Opt
	writerOpts = append(writerOpts, OptCopy())
	// Add the same artwork twice to test multiple artwork support
	writerOpts = append(writerOpts, OptMetadata(NewMetadata(MetaArtwork, artworkData)))
	writerOpts = append(writerOpts, OptMetadata(NewMetadata(MetaArtwork, artworkData)))

	// Add streams from input
	streamMap := make(map[int]int)
	outputIndex := 0
	for _, stream := range reader.Streams(media.ANY) {
		par := &Par{AVCodecParameters: *stream.CodecPar()}
		writerOpts = append(writerOpts, OptStream(0, par))
		streamMap[stream.Index()] = outputIndex
		outputIndex++
	}

	// Create writer
	writer, err := Create(outputFile, writerOpts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Copy packets (both artworks will be written automatically on first packet)
	packetCount := 0
	ctx := context.Background()
	err = reader.Decode(ctx, func(stream int, pkt *Packet) error {
		if newStream, ok := streamMap[stream]; ok {
			pkt.AVPacket.SetStreamIndex(newStream)
		}
		packetCount++
		return writer.Write(pkt)
	})

	assert.NoError(err)
	assert.Greater(packetCount, 0)
	t.Logf("Remuxed %d packets with multiple artworks to %s", packetCount, outputFile)

	// Verify output file exists
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
}

func Test_encode_sync_new_frames_mp4(t *testing.T) {
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "sync_new_frames.mp4")

	// Create video parameters (640x480, 25fps, H.264)
	videoPar, err := NewVideoPar("yuv420p", "640x480", 25.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with video stream
	writer, err := Create(outputFile, OptStream(0, videoPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Encode 50 frames, creating a new frame each time
	for i := 0; i < 50; i++ {
		frame, err := NewFrame(videoPar)
		if !assert.NoError(err) {
			return
		}
		if err := frame.AllocateBuffers(); !assert.NoError(err) {
			frame.Close()
			return
		}
		fillVideoBlackYUV420P(frame)
		frame.SetPts(int64(i))
		(*ff.AVFrame)(frame).SetDuration(1)

		if err := writer.EncodeFrame(0, frame); !assert.NoError(err) {
			frame.Close()
			return
		}
		frame.Close()
	}

	// Flush encoder
	if err := writer.EncodeFrame(0, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
}

func Test_encode_sync_via_encodeframes_mp4(t *testing.T) {
	// Fixed: encoder now leaves packet ownership to muxer
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "sync_via_encodeframes.mp4")

	// Create video parameters (640x480, 25fps, H.264)
	videoPar, err := NewVideoPar("yuv420p", "640x480", 25.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with video stream
	writer, err := Create(outputFile, OptStream(0, videoPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Create channel
	frames := make(chan *Frame, 10)

	// Feed channel in background
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(frames)
		for i := 0; i < 50; i++ {
			frame, err := NewFrame(videoPar)
			if err != nil {
				return
			}
			if err := frame.AllocateBuffers(); err != nil {
				frame.Close()
				return
			}
			fillVideoBlackYUV420P(frame)
			frame.SetPts(int64(i))
			(*ff.AVFrame)(frame).SetDuration(1)
			frames <- frame
		}
	}()

	// Encode using EncodeFrames
	err = writer.EncodeFrames(0, frames)
	assert.NoError(err)

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME INTERLEAVING

// Test encoding multiple streams with proper temporal interleaving.
// This demonstrates the recommended approach: compare PTS values across streams
// and encode the frame with the earliest timestamp first.
func Test_encode_multiple_streams_interleaved_mp4(t *testing.T) {
	assert := assert.New(t)

	outputFile := filepath.Join(t.TempDir(), "multi_stream_interleaved.mp4")

	// Create video parameters (720p, 30fps)
	videoPar, err := NewVideoPar("yuv420p", "1280x720", 30.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create audio parameters (stereo, 44.1kHz, AAC)
	audioPar, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create writer with both streams
	writer, err := Create(outputFile, OptStream(0, videoPar), OptStream(0, audioPar))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer writer.Close()

	// Create frames
	videoFrame, err := NewFrame(videoPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer videoFrame.Close()
	if err := videoFrame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	audioFrame, err := NewFrame(audioPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer audioFrame.Close()
	(*ff.AVFrame)(audioFrame).SetNumSamples(1024)
	if err := audioFrame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	// Calculate frame counts
	numVideoFrames := 60 // 2 seconds at 30fps
	samplesPerAudioFrame := 1024
	totalAudioSamples := int64(numVideoFrames) * 44100 / 30
	numAudioFrames := int(totalAudioSamples / int64(samplesPerAudioFrame))

	// Track current frame indices
	videoIndex := 0
	audioIndex := 0
	videoSampleIndex := int64(0)
	audioSampleIndex := int64(0)

	// Get stream timebase information for comparison
	videoTimebase := writer.encoders[0].stream.TimeBase()
	audioTimebase := writer.encoders[1].stream.TimeBase()

	// Helper to convert PTS to seconds for comparison
	ptsToSeconds := func(pts int64, timebase ff.AVRational) float64 {
		return float64(pts) * float64(timebase.Num()) / float64(timebase.Den())
	}

	// Interleave frames by comparing timestamps
	for videoIndex < numVideoFrames || audioIndex < numAudioFrames {
		// Calculate current timestamps in seconds
		var videoTime, audioTime float64
		hasVideo := videoIndex < numVideoFrames
		hasAudio := audioIndex < numAudioFrames

		if hasVideo {
			videoTime = ptsToSeconds(videoSampleIndex, videoTimebase)
		}
		if hasAudio {
			audioTime = ptsToSeconds(audioSampleIndex, audioTimebase)
		}

		// Decide which stream to encode next based on timestamps
		if hasVideo && (!hasAudio || videoTime <= audioTime) {
			// Encode video frame
			videoFrame.SetPts(videoSampleIndex)
			if err := writer.EncodeFrame(0, videoFrame); !assert.NoError(err) {
				t.FailNow()
			}
			videoIndex++
			videoSampleIndex++
		} else if hasAudio {
			// Encode audio frame
			audioFrame.SetPts(audioSampleIndex)
			if err := writer.EncodeFrame(1, audioFrame); !assert.NoError(err) {
				t.FailNow()
			}
			audioIndex++
			audioSampleIndex += int64(samplesPerAudioFrame)
		}
	}

	// Flush both encoders
	if err := writer.EncodeFrame(0, nil); !assert.NoError(err) {
		t.FailNow()
	}
	if err := writer.EncodeFrame(1, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Close writer
	if err := writer.Close(); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output
	info, err := os.Stat(outputFile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0))
	t.Logf("Created interleaved multi-stream video: %s (%d bytes, %d video frames, %d audio frames)",
		outputFile, info.Size(), numVideoFrames, numAudioFrames)
}

// Test demonstrating why interleaving matters: this shows the difference between
// sending all video first vs properly interleaved.
func Test_encode_interleaving_comparison(t *testing.T) {
	assert := assert.New(t)

	// Test 1: All video first, then all audio (not optimal)
	outputFile1 := filepath.Join(t.TempDir(), "video_first.mp4")
	createVideoAudioFile(t, outputFile1, false)
	info1, _ := os.Stat(outputFile1)

	// Test 2: Properly interleaved (optimal)
	outputFile2 := filepath.Join(t.TempDir(), "interleaved.mp4")
	createVideoAudioFile(t, outputFile2, true)
	info2, _ := os.Stat(outputFile2)

	t.Logf("Video-first approach: %d bytes", info1.Size())
	t.Logf("Interleaved approach: %d bytes", info2.Size())
	t.Logf("Both approaches work, but interleaved is better for streaming and reduces muxer buffering")

	// Both should produce valid files
	assert.Greater(info1.Size(), int64(0))
	assert.Greater(info2.Size(), int64(0))
}

////////////////////////////////////////////////////////////////////////////////
// HELPER FUNCTIONS

func createVideoAudioFile(t *testing.T, outputFile string, interleave bool) {
	assert := assert.New(t)

	videoPar, _ := NewVideoPar("yuv420p", "640x480", 25.0)
	audioPar, _ := NewAudioPar("fltp", "stereo", 44100)

	writer, err := Create(outputFile, OptStream(0, videoPar), OptStream(0, audioPar))
	if !assert.NoError(err) {
		return
	}
	defer writer.Close()

	videoFrame, _ := NewFrame(videoPar)
	defer videoFrame.Close()
	videoFrame.AllocateBuffers()

	audioFrame, _ := NewFrame(audioPar)
	defer audioFrame.Close()
	(*ff.AVFrame)(audioFrame).SetNumSamples(1024)
	audioFrame.AllocateBuffers()

	numVideoFrames := 30
	numAudioFrames := 34 // ~1.2 seconds of audio at 44.1kHz

	if interleave {
		// Properly interleaved: alternate based on timestamps
		videoTimebase := writer.encoders[0].stream.TimeBase()
		audioTimebase := writer.encoders[1].stream.TimeBase()

		videoIdx, audioIdx := 0, 0
		videoSamples, audioSamples := int64(0), int64(0)

		for videoIdx < numVideoFrames || audioIdx < numAudioFrames {
			videoTime := float64(videoSamples) * float64(videoTimebase.Num()) / float64(videoTimebase.Den())
			audioTime := float64(audioSamples) * float64(audioTimebase.Num()) / float64(audioTimebase.Den())

			if videoIdx < numVideoFrames && (audioIdx >= numAudioFrames || videoTime <= audioTime) {
				videoFrame.SetPts(videoSamples)
				writer.EncodeFrame(0, videoFrame)
				videoIdx++
				videoSamples++
			} else if audioIdx < numAudioFrames {
				audioFrame.SetPts(audioSamples)
				writer.EncodeFrame(1, audioFrame)
				audioIdx++
				audioSamples += 1024
			}
		}
	} else {
		// Non-interleaved: all video first, then all audio
		for i := 0; i < numVideoFrames; i++ {
			videoFrame.SetPts(int64(i))
			writer.EncodeFrame(0, videoFrame)
		}
		writer.EncodeFrame(0, nil) // Flush video

		for i := 0; i < numAudioFrames; i++ {
			audioFrame.SetPts(int64(i * 1024))
			writer.EncodeFrame(1, audioFrame)
		}
		writer.EncodeFrame(1, nil) // Flush audio
	}

	writer.Close()
}
