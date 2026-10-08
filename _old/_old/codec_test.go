package manager_test

import (
	"strings"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestListCodecs_All(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListCodecs(ctx, schema.ListCodecRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one codec")
	}
}

func TestListCodecs_FilterName(t *testing.T) {
	m, ctx := test.Begin(t)

	all, err := m.ListCodecs(ctx, schema.ListCodecRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("expected codecs for name filter test")
	}

	name := all[0].Name()
	resp, err := m.ListCodecs(ctx, schema.ListCodecRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatalf("expected at least one codec for name %q", name)
	}

	for i, codec := range resp {
		if !strings.Contains(codec.Name(), name) {
			t.Fatalf("codec[%d] name=%q does not contain %q", i, codec.Name(), name)
		}
	}
}

func TestListCodecs_FilterType(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListCodecs(ctx, schema.ListCodecRequest{Type: "audio"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one audio codec")
	}

	for i, codec := range resp {
		mt := codec.Type()
		if mt == nil || mt.String() != "audio" {
			t.Fatalf("codec[%d] media type=%v, want audio", i, mt)
		}
	}
}

func TestListCodecs_FilterEncoder(t *testing.T) {
	m, ctx := test.Begin(t)

	isEncoder := true
	resp, err := m.ListCodecs(ctx, schema.ListCodecRequest{IsEncoder: &isEncoder})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one encoder codec")
	}

	for i, codec := range resp {
		if !ff.AVCodec_is_encoder(codec.AVCodec) {
			t.Fatalf("codec[%d] (%q) is not an encoder", i, codec.Name())
		}
	}
}
