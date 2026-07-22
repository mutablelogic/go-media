package manager_test

import (
	"testing"

	// Packages
	test "github.com/mutablelogic/go-media/gomedia/test"
	profile "github.com/mutablelogic/go-media/profile/schema"
)

func monoFloatProfile(t *testing.T) *profile.AudioProfile {
	t.Helper()
	p, err := profile.NewAudioProfile("pcm_f32le")
	if err != nil {
		t.Fatalf("NewAudioProfile(pcm_f32le): %v", err)
	}
	if err := p.Set(profile.OptionSampleRate, uint64(44100)); err != nil {
		t.Fatalf("Set(sample_rate): %v", err)
	}
	if err := p.Set(profile.OptionSampleFormat, "flt"); err != nil {
		t.Fatalf("Set(sample_format): %v", err)
	}
	if err := p.Set(profile.OptionChannelLayout, "mono"); err != nil {
		t.Fatalf("Set(channel_layout): %v", err)
	}
	return p
}

func TestCreateGetDeleteSource(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	src, err := m.CreateSource(ctx, "tone", monoFloatProfile(t))
	if err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
	if src == nil {
		t.Fatal("CreateSource: expected non-nil source")
	}
	defer m.DeleteSource(ctx, "tone")

	got, err := m.GetSource(ctx, "tone")
	if err != nil {
		t.Fatalf("GetSource: %v", err)
	}
	if got != src {
		t.Fatal("GetSource: expected the same source returned by CreateSource")
	}

	if err := m.DeleteSource(ctx, "tone"); err != nil {
		t.Fatalf("DeleteSource: %v", err)
	}
	if _, err := m.GetSource(ctx, "tone"); err == nil {
		t.Fatal("GetSource: expected error after DeleteSource")
	}
}

func TestCreateSource_EmptyName(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	if _, err := m.CreateSource(ctx, "", monoFloatProfile(t)); err == nil {
		t.Fatal("CreateSource: expected error for empty name")
	}
}

func TestCreateSource_Duplicate(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	if _, err := m.CreateSource(ctx, "dup", monoFloatProfile(t)); err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
	defer m.DeleteSource(ctx, "dup")

	if _, err := m.CreateSource(ctx, "dup", monoFloatProfile(t)); err == nil {
		t.Fatal("CreateSource: expected error for duplicate name")
	}
}

func TestGetSource_NotFound(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	if _, err := m.GetSource(ctx, "missing"); err == nil {
		t.Fatal("GetSource: expected error for unknown name")
	}
}

func TestDeleteSource_NotFound(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	if err := m.DeleteSource(ctx, "missing"); err == nil {
		t.Fatal("DeleteSource: expected error for unknown name")
	}
}
