package reader_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

//////////////////////////////////////////////////////////////////////////////
// HELPERS

func audioProfile(t *testing.T) *profile.AudioProfile {
	t.Helper()
	p, err := profile.NewAudioProfile("aac")
	if err != nil {
		t.Fatalf("NewAudioProfile(aac): %v", err)
	}
	if err := p.Set(profile.OptionSampleRate, uint64(44100)); err != nil {
		t.Fatalf("Set(sample_rate): %v", err)
	}
	if err := p.Set(profile.OptionSampleFormat, "fltp"); err != nil {
		t.Fatalf("Set(sample_format): %v", err)
	}
	if err := p.Set(profile.OptionChannelLayout, "stereo"); err != nil {
		t.Fatalf("Set(channel_layout): %v", err)
	}
	return p
}

func silentFrame(t *testing.T, streamID, numSamples int) *frame.Frame {
	t.Helper()

	f, err := frame.NewFrame(streamID)
	if err != nil {
		t.Fatalf("NewFrame: %v", err)
	}

	f.SetSampleFormat(ff.AVUtil_get_sample_fmt("fltp"))
	f.SetSampleRate(44100)

	var ch ff.AVChannelLayout
	if err := ff.AVUtil_channel_layout_from_string(&ch, "stereo"); err != nil {
		t.Fatalf("AVUtil_channel_layout_from_string: %v", err)
	}
	if err := f.SetChannelLayout(ch); err != nil {
		t.Fatalf("SetChannelLayout: %v", err)
	}
	f.SetNumSamples(numSamples)

	if err := f.AllocateBuffers(); err != nil {
		t.Fatalf("AllocateBuffers: %v", err)
	}

	for plane := 0; plane < f.NumChannels(); plane++ {
		samples := f.Float32(plane)
		for i := range samples {
			samples[i] = 0
		}
	}

	return f
}

// writeSampleFile encodes ~2.3s of silent AAC audio to path, using the
// already-tested writer package, so reader tests have a real file to open.
func writeSampleFile(t *testing.T, path string) {
	t.Helper()

	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(&url.URL{Path: path}, output, writer.WithProfile(0, audioProfile(t)))
	if err != nil {
		t.Fatalf("writer.Create: %v", err)
	}

	numSamples := w.FrameSize(0)
	if numSamples == 0 {
		numSamples = 1024
	}

	const numFrames = 100
	for i := 0; i < numFrames; i++ {
		f := silentFrame(t, 0, numSamples)
		f.SetPts(int64(i * numSamples))
		if err := w.Encode(f); err != nil {
			f.Close()
			t.Fatalf("Encode: %v", err)
		}
		f.Close()
	}

	if err := w.Flush(0); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close (writer): %v", err)
	}
}

func sampleFilePath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.mp4")
	writeSampleFile(t, path)
	return path
}

// artworkMeta is a minimal gomedia.Metadata implementation carrying artwork
// (image) data under the gomedia.MetaArtwork key.
type artworkMeta struct {
	data []byte
}

func (a artworkMeta) Key() string        { return gomedia.MetaArtwork }
func (a artworkMeta) Value() string      { return "" }
func (a artworkMeta) Bytes() []byte      { return a.data }
func (a artworkMeta) Image() image.Image { return nil }
func (a artworkMeta) Any() any           { return a.data }

func testJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return buf.Bytes()
}

// writeSampleFileWithArtwork is like writeSampleFile, but also embeds
// artwork as attached-pic metadata, for round-tripping through Metadata().
func writeSampleFileWithArtwork(t *testing.T, path string, artwork []byte) {
	t.Helper()

	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(&url.URL{Path: path}, output,
		writer.WithProfile(0, audioProfile(t)),
		writer.WithMetadata(artworkMeta{data: artwork}),
	)
	if err != nil {
		t.Fatalf("writer.Create: %v", err)
	}

	numSamples := w.FrameSize(0)
	if numSamples == 0 {
		numSamples = 1024
	}

	const numFrames = 10
	for i := 0; i < numFrames; i++ {
		f := silentFrame(t, 0, numSamples)
		f.SetPts(int64(i * numSamples))
		if err := w.Encode(f); err != nil {
			f.Close()
			t.Fatalf("Encode: %v", err)
		}
		f.Close()
	}

	if err := w.Flush(0); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close (writer): %v", err)
	}
}

//////////////////////////////////////////////////////////////////////////////
// TESTS

func TestOpen(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	if d := r.Duration(); d <= 0 {
		t.Fatalf("Duration() = %v, want > 0", d)
	}
}

func TestOpen_NotFound(t *testing.T) {
	if _, err := reader.Open(filepath.Join(t.TempDir(), "does-not-exist.mp4")); err == nil {
		t.Fatal("Open: expected error for nonexistent file")
	}
}

func TestNewReader(t *testing.T) {
	data, err := os.ReadFile(sampleFilePath(t))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	r, err := reader.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	if d := r.Duration(); d <= 0 {
		t.Fatalf("Duration() = %v, want > 0", d)
	}
}

// Exercises the open()/find_stream_info failure path directly (this is
// where a *Reader created via NewReader used to leak its AVIOContext on
// failure, since only r.input was freed, never r.avio).
func TestNewReader_InvalidData(t *testing.T) {
	if _, err := reader.NewReader(bytes.NewReader([]byte("not a real media file"))); err == nil {
		t.Fatal("NewReader: expected error for invalid data")
	}
}

func TestNewReader_EmptyData(t *testing.T) {
	if _, err := reader.NewReader(bytes.NewReader(nil)); err == nil {
		t.Fatal("NewReader: expected error for empty data")
	}
}

func TestReader_Close_Idempotent(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close (first): %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close (second): %v", err)
	}
}

// Regression test: Duration() used to dereference r.input directly, which
// would panic (nil pointer) once Close() had nilled it out.
func TestReader_Duration_AfterClose(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if d := r.Duration(); d != 0 {
		t.Fatalf("Duration() after Close = %v, want 0", d)
	}
}

// Regression test: Seek() used to dereference r.input directly, which would
// panic (nil pointer) once Close() had nilled it out.
func TestReader_Seek_AfterClose(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := r.Seek(0, 0*time.Second); err == nil {
		t.Fatal("Seek after Close: expected error")
	}
}

func TestReader_Seek(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	if err := r.Seek(0, 500*time.Millisecond); err != nil {
		t.Fatalf("Seek: %v", err)
	}
}

func TestReader_Seek_InvalidStream(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	if err := r.Seek(99, 0*time.Second); err == nil {
		t.Fatal("Seek: expected error for invalid stream index")
	}
}

// Regression test: Metadata() called with no keys must still include
// artwork — the artwork block used to only fire when "artwork" was
// explicitly requested, silently omitting it from the "return everything"
// (no filter) case.
func TestReader_Metadata_Artwork(t *testing.T) {
	artwork := testJPEG(t)
	path := filepath.Join(t.TempDir(), "sample_with_artwork.mp4")
	writeSampleFileWithArtwork(t, path, artwork)

	r, err := reader.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	entries := r.Metadata()

	var found gomedia.Metadata
	count := 0
	for _, e := range entries {
		if e.Key() == gomedia.MetaArtwork {
			found = e
			count++
		}
	}
	if found == nil {
		t.Fatal("Metadata(): expected an artwork entry when called with no keys")
	}
	if count != 1 {
		t.Fatalf("Metadata(): expected exactly 1 artwork entry, got %d", count)
	}
	if !bytes.Equal(found.Bytes(), artwork) {
		t.Fatalf("Metadata(): artwork bytes mismatch, got %d bytes, want %d bytes", len(found.Bytes()), len(artwork))
	}
}

func TestReader_Metadata_FilterByKey(t *testing.T) {
	artwork := testJPEG(t)
	path := filepath.Join(t.TempDir(), "sample_with_artwork.mp4")
	writeSampleFileWithArtwork(t, path, artwork)

	r, err := reader.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	entries := r.Metadata(gomedia.MetaArtwork)
	if len(entries) != 1 {
		t.Fatalf("Metadata(artwork): expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key() != gomedia.MetaArtwork {
		t.Fatalf("Metadata(artwork): got key %q, want %q", entries[0].Key(), gomedia.MetaArtwork)
	}
}

// Regression test: Metadata() used to dereference r.input directly, which
// would panic (nil pointer) once Close() had nilled it out.
func TestReader_Metadata_AfterClose(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if got := r.Metadata(); got != nil {
		t.Fatalf("Metadata() after Close = %v, want nil", got)
	}
}

func TestReader_Streams(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	streams := r.Streams()
	if len(streams) != 1 {
		t.Fatalf("Streams(): expected 1 stream, got %d", len(streams))
	}

	p, ok := streams[0]
	if !ok {
		t.Fatal("Streams(): expected an entry keyed by stream index 0")
	}
	if p.Type() != profile.CodecType(ff.AVMEDIA_TYPE_AUDIO) {
		t.Fatalf("Streams()[0].Type() = %v, want audio", p.Type())
	}
}

// Cover art is demuxed as a video stream (with the attached-pic
// disposition), not as an AVMEDIA_TYPE_ATTACHMENT stream — Streams() must
// exclude it so it isn't returned twice (once here, once via Metadata's
// "artwork" key).
func TestReader_Streams_ExcludesArtwork(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample_with_artwork.mp4")
	writeSampleFileWithArtwork(t, path, testJPEG(t))

	r, err := reader.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	streams := r.Streams()
	if len(streams) != 1 {
		t.Fatalf("Streams(): expected 1 stream (artwork excluded), got %d", len(streams))
	}
	if _, ok := streams[0]; !ok {
		t.Fatal("Streams(): expected the audio stream keyed by index 0")
	}
}

func TestReader_Streams_AfterClose(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if got := r.Streams(); got != nil {
		t.Fatalf("Streams() after Close = %v, want nil", got)
	}
}

// Demonstrates the intended usage pattern documented on Streams(): feeding
// its output straight into writer.WithProfile to remux into a new file.
func TestReader_Streams_Remux(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	var opts []writer.Opt
	for i, p := range r.Streams() {
		opts = append(opts, writer.WithProfile(i, p))
	}
	if len(opts) != 1 {
		t.Fatalf("expected 1 WithProfile option, got %d", len(opts))
	}

	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(&url.URL{Path: filepath.Join(t.TempDir(), "remux.mp4")}, output, opts...)
	if err != nil {
		t.Fatalf("writer.Create with Streams() profiles: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestReader_Decode_PacketFn(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	var packets int
	err = r.Decode(context.Background(), func(*ff.AVPacket) error {
		packets++
		return nil
	}, nil)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if packets == 0 {
		t.Fatal("expected at least one packet")
	}
}

// Regression test: packetfn returning io.EOF should end Decode early without
// an error, matching the documented "stop early, no error" contract.
func TestReader_Decode_PacketFn_EOF(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	var packets int
	err = r.Decode(context.Background(), func(*ff.AVPacket) error {
		packets++
		return io.EOF
	}, nil)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if packets != 1 {
		t.Fatalf("expected exactly 1 packet before stopping, got %d", packets)
	}
}

func TestReader_Decode_NilCallbacks(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	if err := r.Decode(context.Background(), nil, nil); err == nil {
		t.Fatal("Decode: expected error when packetfn and decoder are both nil")
	}
}

// Regression test: the reader-is-closed check must run inside the same
// locked section Close() uses, so a race between the two can't slip a nil
// r.input past the check.
func TestReader_Decode_AfterClose(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err = r.Decode(context.Background(), func(*ff.AVPacket) error { return nil }, nil)
	if err == nil {
		t.Fatal("Decode after Close: expected error")
	}
}

func TestReader_Decode_ContextCancelled(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var packets int
	err = r.Decode(ctx, func(*ff.AVPacket) error {
		packets++
		return nil
	}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Decode: got error %v, want context.Canceled", err)
	}
	if packets != 0 {
		t.Fatalf("expected 0 packets read before cancellation was observed, got %d", packets)
	}
}
