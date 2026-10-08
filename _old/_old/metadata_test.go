package manager_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Packages
	test "github.com/mutablelogic/go-media/gomedia/test"

	// Metadata handlers
	_ "github.com/mutablelogic/go-media/metadata/application"
	_ "github.com/mutablelogic/go-media/metadata/audio"
	_ "github.com/mutablelogic/go-media/metadata/image"
	_ "github.com/mutablelogic/go-media/metadata/video"
)

type namedNonSeeker struct {
	r    *bufio.Reader
	name string
}

func (n *namedNonSeeker) Read(p []byte) (int, error) {
	return n.r.Read(p)
}

func (n *namedNonSeeker) Name() string {
	return n.name
}

func testFilePath(t *testing.T, file string) string {
	t.Helper()
	return filepath.Join("..", "..", "etc", "test", file)
}

func TestGetMetadata_All(t *testing.T) {
	m, ctx := test.Begin(t)

	f, err := os.Open(testFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var warn error
	meta, err := m.GetMetadata(ctx, f, "", &warn)
	if err != nil {
		t.Fatal(err)
	}
	if warn != nil {
		t.Fatalf("unexpected warning: %v", warn)
	}
	if meta.ContentType == "" {
		t.Fatal("expected non-empty content type")
	}
	if len(meta.Meta) == 0 {
		t.Fatal("expected metadata entries")
	}
}

func TestGetMetadata_FilterNamespace(t *testing.T) {
	m, ctx := test.Begin(t)

	f, err := os.Open(testFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	meta, err := m.GetMetadata(ctx, f, "audio:", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(meta.Meta) == 0 {
		t.Fatal("expected filtered metadata entries")
	}
	for i, item := range meta.Meta {
		if !strings.HasPrefix(item.Key(), "audio:") {
			t.Fatalf("meta[%d] key=%q does not match filter %q", i, item.Key(), "audio:")
		}
	}
}

func TestGetMetadata_NonSeekerReader(t *testing.T) {
	m, ctx := test.Begin(t)

	f, err := os.Open(testFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	namedReader := &namedNonSeeker{r: bufio.NewReader(f), name: f.Name()}

	meta, err := m.GetMetadata(ctx, namedReader, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if meta.ContentType == "" {
		t.Fatal("expected non-empty content type")
	}
	if len(meta.Meta) == 0 {
		t.Fatal("expected metadata entries from non-seeker reader")
	}
}
