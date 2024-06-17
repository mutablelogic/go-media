package ffmpeg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_swresample_core_000(t *testing.T) {
	assert := assert.New(t)
	ctx := SWResample_alloc()
	assert.NotNil(ctx)
	SWResample_free(ctx)
}
