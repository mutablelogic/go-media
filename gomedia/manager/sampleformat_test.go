package manager_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestListSampleFormats_All(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListSampleFormats(ctx, schema.ListSampleFormatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one sample format")
	}
}

func TestListSampleFormats_FilterName(t *testing.T) {
	m, ctx := test.Begin(t)

	all, err := m.ListSampleFormats(ctx, schema.ListSampleFormatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("expected sample formats for name filter test")
	}

	name := ff.AVUtil_get_sample_fmt_name(all[0].AVSampleFormat)
	resp, err := m.ListSampleFormats(ctx, schema.ListSampleFormatRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatalf("expected at least one sample format for name %q", name)
	}

	for i, sf := range resp {
		got := ff.AVUtil_get_sample_fmt_name(sf.AVSampleFormat)
		if got != name {
			t.Fatalf("samplefmt[%d] name=%q, want %q", i, got, name)
		}
	}
}

func TestListSampleFormats_FilterIsPlanar(t *testing.T) {
	m, ctx := test.Begin(t)

	isPlanar := true
	resp, err := m.ListSampleFormats(ctx, schema.ListSampleFormatRequest{IsPlanar: &isPlanar})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one planar sample format")
	}

	for i, sf := range resp {
		if !ff.AVUtil_sample_fmt_is_planar(sf.AVSampleFormat) {
			t.Fatalf("samplefmt[%d] (%q) is not planar", i, ff.AVUtil_get_sample_fmt_name(sf.AVSampleFormat))
		}
	}
}
