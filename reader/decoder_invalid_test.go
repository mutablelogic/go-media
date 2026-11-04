package reader

import (
	"testing"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
)

// newInvalidDecoder returns a decoder which counts the packets its streams'
// codecs reject, as it does once it's opened
func newInvalidDecoder(t *testing.T) *Decoder {
	t.Helper()
	d, err := NewDecoder(func(frame.Frame) error { return nil })
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}
	d.invalid = make(map[int]int)
	return d
}

// A stream's rejected packets are skipped, until the maxInvalidPackets'th in
// a row, which fails decoding
func TestDecoder_SkipInvalid_Limit(t *testing.T) {
	d := newInvalidDecoder(t)
	for i := 1; i < maxInvalidPackets; i++ {
		if !d.skipInvalid(0) {
			t.Fatalf("rejected packet %d isn't skipped, want it skipped", i)
		}
	}
	if d.skipInvalid(0) {
		t.Fatalf("rejected packet %d is skipped, want decoding to fail", maxInvalidPackets)
	}
}

// A packet which is accepted resets the stream's count, so rejected packets
// which aren't in a row don't fail decoding
func TestDecoder_SkipInvalid_Reset(t *testing.T) {
	d := newInvalidDecoder(t)
	for i := 1; i < maxInvalidPackets; i++ {
		if !d.skipInvalid(0) {
			t.Fatalf("rejected packet %d isn't skipped, want it skipped", i)
		}
	}
	d.validPacket(0)
	for i := 1; i < maxInvalidPackets; i++ {
		if !d.skipInvalid(0) {
			t.Fatalf("rejected packet %d after an accepted one isn't skipped, want it skipped", i)
		}
	}
	if d.skipInvalid(0) {
		t.Fatalf("rejected packet %d after an accepted one is skipped, want decoding to fail", maxInvalidPackets)
	}
}

// Each stream's rejected packets are counted on their own, so another
// stream's don't count towards its limit, and an accepted packet of
// another stream doesn't reset it
func TestDecoder_SkipInvalid_Streams(t *testing.T) {
	d := newInvalidDecoder(t)
	for i := 1; i < maxInvalidPackets; i++ {
		if !d.skipInvalid(0) || !d.skipInvalid(1) {
			t.Fatalf("rejected packet %d of a stream isn't skipped, want it skipped", i)
		}
	}
	d.validPacket(1)
	if !d.skipInvalid(1) {
		t.Fatal("rejected packet of stream 1 after an accepted one isn't skipped, want it skipped")
	}
	if d.skipInvalid(0) {
		t.Fatalf("rejected packet %d of stream 0 is skipped, want decoding to fail", maxInvalidPackets)
	}
}
