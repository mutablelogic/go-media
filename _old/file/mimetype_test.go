package file_test

import (
	"os"
	"testing"

	// Packages
	"github.com/mutablelogic/go-media/pkg/file"
	"github.com/stretchr/testify/assert"
)

const (
	SAMPLE_MP4 = "../../etc/test/sample.mp4"
)

func Test_mimetype_000(t *testing.T) {
	assert := assert.New(t)
	bytes, err := os.ReadFile(SAMPLE_MP4)
	assert.NoError(err)
	mimetype, ext, err := file.MimeType(bytes)
	assert.NoError(err)
	t.Log(mimetype)
	t.Log(ext)
}
