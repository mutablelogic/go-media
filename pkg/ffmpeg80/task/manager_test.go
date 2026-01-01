package task_test

import (
	"os"
	"path/filepath"
	"testing"

	// Packages
	"github.com/mutablelogic/go-media/pkg/ffmpeg80/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDataPath returns the path to the test data directory
func testDataPath(t *testing.T) string {
	// Find the project root by looking for go.mod
	cwd, err := os.Getwd()
	require.NoError(t, err)

	// Walk up to find the go.mod file
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return filepath.Join(cwd, "etc", "test")
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			t.Fatal("could not find project root")
		}
		cwd = parent
	}
}

func TestNewManager(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)
	assert.NotNil(t, m)
}
