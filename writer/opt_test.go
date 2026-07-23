package writer

import (
	"image"
	"testing"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
)

//////////////////////////////////////////////////////////////////////////////
// HELPERS

func testProfile(t *testing.T) *profile.AudioProfile {
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

// stubMetadata is a minimal gomedia.Metadata implementation for testing
// WithMetadata in isolation.
type stubMetadata struct {
	key, value string
}

func (m stubMetadata) Key() string        { return m.key }
func (m stubMetadata) Value() string      { return m.value }
func (m stubMetadata) Bytes() []byte      { return []byte(m.value) }
func (m stubMetadata) Image() image.Image { return nil }
func (m stubMetadata) Any() any           { return m.value }

//////////////////////////////////////////////////////////////////////////////
// TESTS

func TestWithProfile_AutoIndex(t *testing.T) {
	var o opts
	p0, p1, p2 := testProfile(t), testProfile(t), testProfile(t)

	if err := o.apply(WithProfile(0, p0), WithProfile(0, p1), WithProfile(0, p2)); err != nil {
		t.Fatalf("apply: %v", err)
	}

	for id, want := range map[int]*profile.AudioProfile{0: p0, 1: p1, 2: p2} {
		if o.streams[id] != profile.Profile(want) {
			t.Fatalf("stream %d: got %v, want %v", id, o.streams[id], want)
		}
	}
}

func TestWithProfile_ExplicitIndex(t *testing.T) {
	var o opts
	p := testProfile(t)

	if err := o.apply(WithProfile(5, p)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if o.streams[5] != profile.Profile(p) {
		t.Fatalf("stream 5: got %v, want %v", o.streams[5], p)
	}
	if len(o.streams) != 1 {
		t.Fatalf("expected exactly 1 stream, got %d", len(o.streams))
	}
}

func TestWithProfile_AutoIndexFillsGap(t *testing.T) {
	var o opts
	p0, p5, pGap := testProfile(t), testProfile(t), testProfile(t)

	// Index 0 is auto-assigned, index 5 is explicit, then a further auto
	// call should fill the gap at 1 rather than continue on from 5.
	if err := o.apply(WithProfile(0, p0), WithProfile(5, p5), WithProfile(0, pGap)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if o.streams[1] != profile.Profile(pGap) {
		t.Fatalf("stream 1: got %v, want %v", o.streams[1], pGap)
	}
}

func TestWithProfile_NegativeIndex(t *testing.T) {
	var o opts
	if err := o.apply(WithProfile(-1, testProfile(t))); err == nil {
		t.Fatal("apply: expected error for negative stream index")
	}
}

func TestWithProfile_NilProfile(t *testing.T) {
	var o opts
	if err := o.apply(WithProfile(1, nil)); err == nil {
		t.Fatal("apply: expected error for nil profile")
	}
}

func TestWithProfile_DuplicateIndex(t *testing.T) {
	var o opts
	if err := o.apply(WithProfile(3, testProfile(t)), WithProfile(3, testProfile(t))); err == nil {
		t.Fatal("apply: expected error for duplicate stream index")
	}
}

func TestWithMetadata_Appends(t *testing.T) {
	var o opts
	m1, m2 := stubMetadata{"a", "1"}, stubMetadata{"b", "2"}

	if err := o.apply(WithMetadata(m1), WithMetadata(m2)); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(o.metadata) != 2 {
		t.Fatalf("expected 2 metadata entries, got %d", len(o.metadata))
	}
	if o.metadata[0] != gomedia.Metadata(m1) || o.metadata[1] != gomedia.Metadata(m2) {
		t.Fatalf("metadata entries not in append order: %v", o.metadata)
	}
}
