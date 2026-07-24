package reader_test

import (
	"context"
	"io"
	"net/url"
	"path/filepath"
	"testing"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

func TestNewDecoder_NilFn(t *testing.T) {
	if _, err := reader.NewDecoder(nil); err == nil {
		t.Fatal("NewDecoder: expected error for nil callback function")
	}
}

func TestDecoder_Add_Duplicate(t *testing.T) {
	d, err := reader.NewDecoder(func(frame.Frame) error { return nil })
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}

	if err := d.Add(0); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := d.Add(0); err == nil {
		t.Fatal("Add: expected error for duplicate stream index")
	}
}

func TestReader_Decode_WithDecoder(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	var frames int
	d, err := reader.NewDecoder(func(f frame.Frame) error {
		if f.Stream() != 0 {
			t.Fatalf("frame Stream() = %d, want 0", f.Stream())
		}
		af, ok := f.(*frame.AudioFrame)
		if !ok {
			t.Fatalf("frame type = %T, want *frame.AudioFrame", f)
		}
		if af.NumSamples() == 0 {
			t.Fatal("frame NumSamples() = 0, want > 0")
		}
		frames++
		return nil
	})
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}
	if err := d.Add(0); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := r.Decode(context.Background(), nil, d); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if frames == 0 {
		t.Fatal("expected at least one decoded frame")
	}
}

// The packetfn must see every packet regardless of which streams a Decoder
// has registered - only frame decoding is gated by Decoder.Add.
func TestReader_Decode_PacketFnSeesDiscardedStream(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	var frames int
	d, err := reader.NewDecoder(func(frame.Frame) error {
		frames++
		return nil
	})
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}
	// Deliberately not calling d.Add(0): stream 0 is discarded from frame
	// decoding, but packetfn must still see its packets.

	var packets int
	packetfn := func(*ff.AVPacket) error {
		packets++
		return nil
	}

	if err := r.Decode(context.Background(), packetfn, d); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if packets == 0 {
		t.Fatal("expected packetfn to see packets even though the stream was discarded from decoding")
	}
	if frames != 0 {
		t.Fatalf("expected 0 decoded frames for a discarded stream, got %d", frames)
	}
}

// Regression test: a FrameFn returning io.EOF must stop Decode early
// without an error, matching PacketFn's contract.
func TestReader_Decode_FrameFn_EOF(t *testing.T) {
	r, err := reader.Open(sampleFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	var frames int
	d, err := reader.NewDecoder(func(frame.Frame) error {
		frames++
		return io.EOF
	})
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}
	if err := d.Add(0); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := r.Decode(context.Background(), nil, d); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if frames != 1 {
		t.Fatalf("expected exactly 1 frame before stopping, got %d", frames)
	}
}

// End-to-end: encode a subtitle with writer.Encoder into a real container,
// then decode it back with reader.Decoder. FFmpeg handles subtitles via a
// completely separate legacy API on both the encode and decode sides, so
// this is the one path that actually proves the two ends agree - unit
// tests against either side alone can't catch a mismatch between them.
//
// Of the sampled text subtitle codecs (ass, srt, mov_text, webvtt), only
// "ass" opens cleanly against this build of libavcodec (see
// sys/ffmpeg80's subtitle encode tests), so "ass" is used here.
func TestReader_Decode_Subtitle_RoundTrip(t *testing.T) {
	sub, err := profile.NewSubtitleProfile("ass")
	if err != nil {
		t.Skipf("ass encoder not available: %v", err)
	}

	output := profile.OutputWithName("matroska")
	if output == nil {
		t.Fatal("OutputWithName(matroska): nil output")
	}

	path := filepath.Join(t.TempDir(), "sub.mkv")
	w, err := writer.Create(&url.URL{Path: path}, output, writer.WithProfile(0, sub))
	if err != nil {
		t.Fatalf("writer.Create: %v", err)
	}

	// The dialogue field is the ASS event's own comma-separated payload
	// (ReadOrder,Layer,Style,Name,MarginL,MarginR,MarginV,Effect,Text) -
	// see SetASS's docs - which round-trips verbatim through encode/mux/
	// demux/decode, so that's what's asserted below rather than just the
	// trailing "Hello world" text field.
	const dialogue = "0,0,Default,,0,0,0,,Hello world"

	subData := ff.NewSubtitle(0)
	if err := subData.SetASS(dialogue, 0, 1000); err != nil {
		t.Fatalf("SetASS: %v", err)
	}
	sf := frame.NewSubtitleFrame(0, subData)
	if err := w.Encode(sf); err != nil {
		sf.Close()
		t.Fatalf("Encode: %v", err)
	}
	sf.Close()

	if err := w.Close(); err != nil {
		t.Fatalf("Close (writer): %v", err)
	}

	r, err := reader.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer r.Close()

	var frames int
	var gotText string
	d, err := reader.NewDecoder(func(f frame.Frame) error {
		sf, ok := f.(*frame.SubtitleFrame)
		if !ok {
			t.Fatalf("frame type = %T, want *frame.SubtitleFrame", f)
		}
		frames++
		if rects := sf.Rects(); len(rects) > 0 {
			gotText = rects[0].Text()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}
	if err := d.Add(0); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := r.Decode(context.Background(), nil, d); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if frames == 0 {
		t.Fatal("expected at least one decoded subtitle")
	}
	if gotText != dialogue {
		t.Fatalf("decoded subtitle text = %q, want %q", gotText, dialogue)
	}
}
