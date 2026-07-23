package reader_test

import (
	"context"
	"io"
	"testing"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestNewDecoder_NilFn(t *testing.T) {
	if _, err := reader.NewDecoder(nil); err == nil {
		t.Fatal("NewDecoder: expected error for nil callback function")
	}
}

func TestDecoder_Add_Duplicate(t *testing.T) {
	d, err := reader.NewDecoder(func(*frame.Frame) error { return nil })
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
	d, err := reader.NewDecoder(func(f *frame.Frame) error {
		if f.StreamID != 0 {
			t.Fatalf("frame StreamID = %d, want 0", f.StreamID)
		}
		if f.NumSamples() == 0 {
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
	d, err := reader.NewDecoder(func(*frame.Frame) error {
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
	d, err := reader.NewDecoder(func(*frame.Frame) error {
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
