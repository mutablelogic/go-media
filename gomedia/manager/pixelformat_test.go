package manager_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestListPixelFormats_All(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListPixelFormats(ctx, schema.ListPixelFormatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one pixel format")
	}
}

func TestListPixelFormats_FilterName(t *testing.T) {
	m, ctx := test.Begin(t)

	all, err := m.ListPixelFormats(ctx, schema.ListPixelFormatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("expected pixel formats for name filter test")
	}

	name := ff.AVUtil_get_pix_fmt_name(all[0].AVPixelFormat)
	resp, err := m.ListPixelFormats(ctx, schema.ListPixelFormatRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatalf("expected at least one pixel format for name %q", name)
	}

	for i, pf := range resp {
		got := ff.AVUtil_get_pix_fmt_name(pf.AVPixelFormat)
		if got != name {
			t.Fatalf("pixfmt[%d] name=%q, want %q", i, got, name)
		}
	}
}

func TestListPixelFormats_FilterNumPlanes(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListPixelFormats(ctx, schema.ListPixelFormatRequest{NumPlanes: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one pixel format with one plane")
	}

	for i, pf := range resp {
		if planes := ff.AVUtil_pix_fmt_count_planes(pf.AVPixelFormat); planes != 1 {
			t.Fatalf("pixfmt[%d] planes=%d, want 1", i, planes)
		}
	}
}
