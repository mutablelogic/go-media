package manager_test

import (
	"testing"

	// Packages
	test "github.com/mutablelogic/go-media/gomedia/test"
)

func TestManagerCreated(t *testing.T) {
	m, _ := test.Begin(t)
	if m == nil {
		t.Fatal("expected shared manager to be initialized")
	}
}
