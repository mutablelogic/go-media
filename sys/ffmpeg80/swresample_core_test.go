package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_swresample_core_000(t *testing.T) {
	assert := assert.New(t)
	ctx := SWResample_alloc()
	assert.NotNil(ctx)
	SWResample_free(ctx)
}
