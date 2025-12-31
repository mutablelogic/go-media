package file_test

import (
	"context"
	"os"
	"testing"

	"github.com/mutablelogic/go-media/pkg/file"
	"github.com/stretchr/testify/assert"
)

func Test_walkfs_000(t *testing.T) {
	assert := assert.New(t)
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}
		t.Log(info.Name())
		t.Log(" root=", root)
		t.Log(" rel=", relpath)
		return nil
	})
	err := walker.Walk(context.Background(), "../..")
	assert.NoError(err)
	t.Log("count=", walker.Count())
}
