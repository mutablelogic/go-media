package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_swresample_opts_000(t *testing.T) {
	in_chlayout := AV_CHANNEL_LAYOUT_STEREO()
	out_chlayout := AV_CHANNEL_LAYOUT_MONO()
	in_format := AV_SAMPLE_FMT_FLTP
	out_format := AV_SAMPLE_FMT_S16

	assert := assert.New(t)
	ctx := SWResample_alloc()
	assert.NotNil(ctx)
	assert.NoError(SWResample_set_opts(ctx, out_chlayout, out_format, 44100, in_chlayout, in_format, 48000))
	assert.NoError(SWResample_init(ctx))
	assert.True(SWResample_is_initialized(ctx))
	SWResample_free(ctx)
}
