package ffmpeg_test

import (
	"os"
	"path/filepath"
	"testing"

	// Pacakge imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
	"github.com/stretchr/testify/assert"
)

const (
	SAMPLE_MP4 = "../../etc/test/sample.mp4"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_avformat_000(t *testing.T) {
	t.Log("avformat_version=", ffmpeg.AVFormat_version())
}

func Test_avformat_001(t *testing.T) {
	t.Log("avformat_configuration=", ffmpeg.AVFormat_configuration())
}

func Test_avformat_002(t *testing.T) {
	t.Log("avformat_license=", ffmpeg.AVFormat_license())
}

func Test_avformat_003(t *testing.T) {
	assert := assert.New(t)
	var opaque uintptr
	for {
		format := ffmpeg.AVFormat_av_muxer_iterate(&opaque)
		if format == nil {
			break
		}
		t.Log("muxer=", format)
		if id := format.DefaultAudioCodec(); id != ffmpeg.AV_CODEC_ID_NONE {
			codec := ffmpeg.AVCodec_find_encoder(id)
			assert.NotNil(codec)
			t.Log("  audio_codec=", codec)
		}
		if id := format.DefaultVideoCodec(); id != ffmpeg.AV_CODEC_ID_NONE {
			codec := ffmpeg.AVCodec_find_encoder(id)
			assert.NotNil(codec)
			t.Log("  video_codec=", codec)
		}
		if id := format.DefaultSubtitleCodec(); id != ffmpeg.AV_CODEC_ID_NONE {
			codec := ffmpeg.AVCodec_find_encoder(id)
			assert.NotNil(codec)
			t.Log("  subtitle_codec=", codec)
		}
	}
}

func Test_avformat_004(t *testing.T) {
	var opaque uintptr
	for {
		format := ffmpeg.AVFormat_av_demuxer_iterate(&opaque)
		if format == nil {
			break
		}
		t.Log("demuxer=", format)
	}
}

func Test_avformat_005(t *testing.T) {
	assert := assert.New(t)
	var ctx *ffmpeg.AVFormatContext
	var dict *ffmpeg.AVDictionary
	ctx, err := ffmpeg.AVFormat_open_input(SAMPLE_MP4, nil, &dict)
	assert.NoError(err)
	assert.NotNil(ctx)
	t.Log(ctx, dict)
	ffmpeg.AVFormat_close_input(&ctx)
	assert.Nil(ctx)
}

func Test_avformat_006(t *testing.T) {
	assert := assert.New(t)
	tmp, err := os.MkdirTemp("", t.Name())
	assert.NoError(err)
	defer os.RemoveAll(tmp)
	ctx, err := ffmpeg.AVFormat_alloc_output_context2(nil, "", filepath.Join(tmp, filepath.Base(SAMPLE_MP4)))
	assert.NoError(err)
	assert.NotNil(ctx)
	t.Log(ctx)
	ffmpeg.AVFormat_free_context(ctx)
}

func Test_avformat_007(t *testing.T) {
	assert := assert.New(t)
	tmp, err := os.MkdirTemp("", t.Name())
	assert.NoError(err)
	defer os.RemoveAll(tmp)
	ctx, err := ffmpeg.AVFormat_alloc_output_context2(nil, "", filepath.Join(tmp, filepath.Base(SAMPLE_MP4)))
	assert.NoError(err)
	assert.NotNil(ctx)
	assert.False(ctx.Output().Format().Is(ffmpeg.AVFMT_NOFILE))
	ioctx, err := ffmpeg.AVFormat_avio_open(ctx.Url(), ffmpeg.AVIO_FLAG_WRITE)
	assert.NoError(err)
	assert.NotNil(ioctx)
	ctx.SetPB(ioctx)
	assert.Equal(ioctx, ctx.PB())
	t.Log(ctx)
	ffmpeg.AVFormat_avio_close(ioctx)
	ffmpeg.AVFormat_free_context(ctx)
}
