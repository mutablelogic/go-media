package manager_test

import (
	"strings"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
)

func TestListFormats_All(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListFormats(ctx, schema.ListFormatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one format")
	}
}

func TestListFormats_FilterName(t *testing.T) {
	m, ctx := test.Begin(t)

	all, err := m.ListFormats(ctx, schema.ListFormatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("expected formats for name filter test")
	}

	name := all[0].Name
	resp, err := m.ListFormats(ctx, schema.ListFormatRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatalf("expected at least one format for name %q", name)
	}

	for i, f := range resp {
		if !strings.Contains(f.Name, name) {
			t.Fatalf("format[%d] name=%q does not contain %q", i, f.Name, name)
		}
	}
}

func TestListFormats_FilterInput(t *testing.T) {
	m, ctx := test.Begin(t)

	isInput := true
	resp, err := m.ListFormats(ctx, schema.ListFormatRequest{IsInput: &isInput})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one input format")
	}

	for i, f := range resp {
		if !f.IsInput {
			t.Fatalf("format[%d] (%q) is not input", i, f.Name)
		}
	}
}

func TestListFormats_FilterOutput(t *testing.T) {
	m, ctx := test.Begin(t)

	isOutput := true
	resp, err := m.ListFormats(ctx, schema.ListFormatRequest{IsOutput: &isOutput})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one output format")
	}

	for i, f := range resp {
		if !f.IsOutput {
			t.Fatalf("format[%d] (%q) is not output", i, f.Name)
		}
	}
}

func TestListFormats_FilterDevice(t *testing.T) {
	m, ctx := test.Begin(t)

	isDevice := true
	resp, err := m.ListFormats(ctx, schema.ListFormatRequest{IsDevice: &isDevice})
	if err != nil {
		t.Fatal(err)
	}

	for i, f := range resp {
		if !f.IsDevice {
			t.Fatalf("format[%d] (%q) is not device", i, f.Name)
		}
	}
}
