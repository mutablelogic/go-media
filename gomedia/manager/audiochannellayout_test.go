package manager_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestListAudioChannelLayouts_All(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListAudioChannelLayouts(ctx, schema.ListAudioChannelLayoutRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one channel layout")
	}
}

func TestListAudioChannelLayouts_FilterNumChannels(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListAudioChannelLayouts(ctx, schema.ListAudioChannelLayoutRequest{NumChannels: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one stereo layout")
	}

	for i, layout := range resp {
		if layout.NumChannels() != 2 {
			t.Fatalf("layout[%d] has %d channels, want 2", i, layout.NumChannels())
		}
	}
}

func TestListAudioChannelLayouts_FilterName(t *testing.T) {
	m, ctx := test.Begin(t)

	all, err := m.ListAudioChannelLayouts(ctx, schema.ListAudioChannelLayoutRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("expected channel layouts for name filter test")
	}

	name, err := ff.AVUtil_channel_layout_describe(all[0].AVChannelLayout)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := m.ListAudioChannelLayouts(ctx, schema.ListAudioChannelLayoutRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatalf("expected at least one layout for name %q", name)
	}

	for i, layout := range resp {
		got, err := ff.AVUtil_channel_layout_describe(layout.AVChannelLayout)
		if err != nil {
			t.Fatal(err)
		}
		if got != name {
			t.Fatalf("layout[%d] name=%q, want %q", i, got, name)
		}
	}
}
