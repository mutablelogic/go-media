package manager_test

import (
	"strings"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
)

func TestListFilters_All(t *testing.T) {
	m, ctx := test.Begin(t)

	resp, err := m.ListFilters(ctx, schema.ListFilterRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatal("expected at least one filter")
	}
}

func TestListFilters_FilterName(t *testing.T) {
	m, ctx := test.Begin(t)

	all, err := m.ListFilters(ctx, schema.ListFilterRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("expected filters for name filter test")
	}

	name := all[0].Name()
	resp, err := m.ListFilters(ctx, schema.ListFilterRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) == 0 {
		t.Fatalf("expected at least one filter for name %q", name)
	}

	for i, filter := range resp {
		if !strings.Contains(filter.Name(), name) {
			t.Fatalf("filter[%d] name=%q does not contain %q", i, filter.Name(), name)
		}
	}
}
