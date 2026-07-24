package schema_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/url"
	"path/filepath"
	"testing"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

//////////////////////////////////////////////////////////////////////////////
// HELPERS

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

func audioTestProfile(t *testing.T) *profile.AudioProfile {
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

// writeSampleFile writes a short AAC audio file, with artwork if provided,
// using the already-tested writer package.
func writeSampleFile(t *testing.T, path string, artwork []byte) {
	t.Helper()

	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	opts := []writer.Opt{writer.WithProfile(0, audioTestProfile(t))}
	if artwork != nil {
		opts = append(opts, writer.WithMetadata(artworkMeta{data: artwork}))
	}

	w, err := writer.Create(&url.URL{Path: path}, output, opts...)
	if err != nil {
		t.Fatalf("writer.Create: %v", err)
	}

	numSamples := w.FrameSize(0)
	if numSamples == 0 {
		numSamples = 1024
	}

	for i := 0; i < 10; i++ {
		f, err := frame.NewAudioFrame(0)
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

// openStreams opens path via the same low-level calls reader.Open uses
// internally (reader.Reader doesn't expose raw *ff.AVStream yet), returning
// the demuxed streams and a cleanup function.
func openStreams(t *testing.T, path string) ([]*ff.AVStream, func()) {
	t.Helper()

	input, err := ff.AVFormat_open_url(path, nil, nil)
	if err != nil {
		t.Fatalf("AVFormat_open_url: %v", err)
	}
	if err := ff.AVFormat_find_stream_info(input, nil); err != nil {
		ff.AVFormat_free_context(input)
		t.Fatalf("AVFormat_find_stream_info: %v", err)
	}

	return input.Streams(), func() { ff.AVFormat_free_context(input) }
}

//////////////////////////////////////////////////////////////////////////////
// TESTS

func TestNewStreamProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.mp4")
	writeSampleFile(t, path, nil)

	streams, cleanup := openStreams(t, path)
	defer cleanup()

	if len(streams) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(streams))
	}

	sp, err := profile.NewStreamProfile(streams[0])
	if err != nil {
		t.Fatalf("NewStreamProfile: %v", err)
	}

	if sp.Index() != 0 {
		t.Fatalf("Index() = %d, want 0", sp.Index())
	}
	if sp.Type() != profile.CodecType(ff.AVMEDIA_TYPE_AUDIO) {
		t.Fatalf("Type() = %v, want audio", sp.Type())
	}
	var zero [16]byte
	if id := sp.UUID(); id != zero {
		t.Fatalf("UUID() = %v, want zero value", id)
	}
	if opts := sp.Options(); opts != nil {
		t.Fatalf("Options() = %s, want nil", opts)
	}

	par := sp.Par()
	if par == nil {
		t.Fatal("Par(): nil")
	}
	if got := par.SampleRate(); got != 44100 {
		t.Fatalf("Par().SampleRate() = %d, want 44100", got)
	}

	if tb := sp.TimeBase(); tb == nil {
		t.Fatal("TimeBase(): nil, want non-nil for a real demuxed stream")
	}
}

func TestNewStreamProfile_Nil(t *testing.T) {
	if _, err := profile.NewStreamProfile(nil); err == nil {
		t.Fatal("NewStreamProfile(nil): expected error")
	}
}

// Cover art is demuxed as a video stream flagged with the attached-pic
// disposition, not as an AVMEDIA_TYPE_ATTACHMENT stream — NewStreamProfile
// must check the disposition explicitly, not just the media type, or
// it would surface cover art as a bogus VideoProfile duplicating what
// Reader.Metadata already returns under the "artwork" key.
func TestNewStreamProfile_AttachedPic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample_with_artwork.mp4")
	writeSampleFile(t, path, testJPEG(t))

	streams, cleanup := openStreams(t, path)
	defer cleanup()

	if len(streams) != 2 {
		t.Fatalf("expected 2 streams (audio + attached-pic), got %d", len(streams))
	}

	var sawAttachedPic bool
	for _, s := range streams {
		if !s.Disposition().Is(ff.AV_DISPOSITION_ATTACHED_PIC) {
			continue
		}
		sawAttachedPic = true
		if _, err := profile.NewStreamProfile(s); err == nil {
			t.Fatal("NewStreamProfile: expected error for attached-pic stream")
		}
	}
	if !sawAttachedPic {
		t.Fatal("expected one of the streams to be an attached-pic stream")
	}
}

// The writer package doesn't currently support setting per-stream metadata
// (only container-level, via WithMetadata), so this sets it directly on the
// demuxed stream to test NewStreamProfile's read side in isolation.
func TestNewStreamProfile_Metadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.mp4")
	writeSampleFile(t, path, nil)

	streams, cleanup := openStreams(t, path)
	defer cleanup()

	if len(streams) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(streams))
	}

	dict := streams[0].Metadata()
	if err := ff.AVUtil_dict_set(dict, "language", "eng", ff.AV_DICT_NONE); err != nil {
		t.Fatalf("AVUtil_dict_set: %v", err)
	}
	streams[0].SetMetadata(dict)

	sp, err := profile.NewStreamProfile(streams[0])
	if err != nil {
		t.Fatalf("NewStreamProfile: %v", err)
	}

	var found gomedia.Metadata
	for _, m := range sp.Metadata() {
		if m.Key() == "language" {
			found = m
		}
	}
	if found == nil {
		t.Fatal("Metadata(): expected a language entry")
	}
	if found.Value() != "eng" {
		t.Fatalf("Metadata(): language = %q, want %q", found.Value(), "eng")
	}
}

// The mp4 muxer auto-populates some stream tags itself (e.g. "language":
// "und", "handler_name": "SoundHandler") even when the caller sets none —
// so this checks for the absence of a tag nothing sets, rather than
// asserting Metadata() is completely empty.
func TestNewStreamProfile_NoMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.mp4")
	writeSampleFile(t, path, nil)

	streams, cleanup := openStreams(t, path)
	defer cleanup()

	sp, err := profile.NewStreamProfile(streams[0])
	if err != nil {
		t.Fatalf("NewStreamProfile: %v", err)
	}
	for _, m := range sp.Metadata() {
		if m.Key() == "title" {
			t.Fatalf("Metadata(): unexpected %q entry", "title")
		}
	}
}
