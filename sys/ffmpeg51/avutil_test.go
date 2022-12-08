package ffmpeg_test

import (
	"testing"

	// Pacakge imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
	assert "github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_avutil_000(t *testing.T) {
	assert := assert.New(t)
	levels := []ffmpeg.AVLogLevel{
		ffmpeg.AV_LOG_QUIET,
		ffmpeg.AV_LOG_PANIC,
		ffmpeg.AV_LOG_FATAL,
		ffmpeg.AV_LOG_ERROR,
		ffmpeg.AV_LOG_WARNING,
		ffmpeg.AV_LOG_INFO,
		ffmpeg.AV_LOG_VERBOSE,
		ffmpeg.AV_LOG_DEBUG,
		ffmpeg.AV_LOG_TRACE,
	}
	for _, level := range levels {
		t.Log("level => ", level)
		ffmpeg.AVUtil_av_log_set_level(level, nil)
		assert.Equal(level, ffmpeg.AVUtil_av_log_get_level())
	}
}

func Test_avutil_001(t *testing.T) {
	var vr ffmpeg.AVLogLevel
	var sr string
	var pr uintptr
	assert := assert.New(t)
	ffmpeg.AVUtil_av_log_set_level(ffmpeg.AV_LOG_QUIET, func(v ffmpeg.AVLogLevel, s string, p uintptr) {
		t.Log("v => ", v)
		t.Log("s => ", s)
		t.Log("p => ", p)
		vr = v
		sr = s
		pr = p
	})
	ffmpeg.AVUtil_av_log(nil, ffmpeg.AV_LOG_QUIET, "test")
	assert.Equal(ffmpeg.AV_LOG_QUIET, vr)
	assert.Equal("test", sr)
	assert.Equal(uintptr(0), pr)
}

func Test_avutil_002(t *testing.T) {
	//assert := assert.New(t)
	errs := []ffmpeg.AVError{
		ffmpeg.AVERROR_BSF_NOT_FOUND,
		ffmpeg.AVERROR_BUG,
		ffmpeg.AVERROR_BUFFER_TOO_SMALL,
		ffmpeg.AVERROR_DECODER_NOT_FOUND,
		ffmpeg.AVERROR_DEMUXER_NOT_FOUND,
		ffmpeg.AVERROR_ENCODER_NOT_FOUND,
		ffmpeg.AVERROR_EOF,
		ffmpeg.AVERROR_EXIT,
		ffmpeg.AVERROR_EXTERNAL,
		ffmpeg.AVERROR_FILTER_NOT_FOUND,
		ffmpeg.AVERROR_INVALIDDATA,
		ffmpeg.AVERROR_MUXER_NOT_FOUND,
		ffmpeg.AVERROR_OPTION_NOT_FOUND,
		ffmpeg.AVERROR_PATCHWELCOME,
		ffmpeg.AVERROR_PROTOCOL_NOT_FOUND,
		ffmpeg.AVERROR_STREAM_NOT_FOUND,
		ffmpeg.AVERROR_BUG2,
		ffmpeg.AVERROR_UNKNOWN,
		ffmpeg.AVERROR_EXPERIMENTAL,
		ffmpeg.AVERROR_INPUT_CHANGED,
		ffmpeg.AVERROR_OUTPUT_CHANGED,
		ffmpeg.AVERROR_HTTP_BAD_REQUEST,
		ffmpeg.AVERROR_HTTP_UNAUTHORIZED,
		ffmpeg.AVERROR_HTTP_FORBIDDEN,
		ffmpeg.AVERROR_HTTP_NOT_FOUND,
		ffmpeg.AVERROR_HTTP_OTHER_4XX,
		ffmpeg.AVERROR_HTTP_SERVER_ERROR,
	}
	for _, err := range errs {
		t.Log("err => ", err)
	}
}

func Test_avutil_003(t *testing.T) {
	assert := assert.New(t)
	dict := ffmpeg.AVUtil_av_dict_new()
	assert.NotNil(dict)
	assert.Equal(0, dict.AVUtil_av_dict_count())
	dict.AVUtil_av_dict_free()
	assert.True(dict.AVUtil_av_dict_context() == nil)
}

func Test_avutil_004(t *testing.T) {
	assert := assert.New(t)
	dict := ffmpeg.AVUtil_av_dict_new()
	assert.NotNil(dict)
	assert.Equal(0, dict.AVUtil_av_dict_count())
	assert.NoError(dict.AVUtil_av_dict_set("a", "b", 0))
	assert.Equal(1, dict.AVUtil_av_dict_count())
	assert.NoError(dict.AVUtil_av_dict_set("b", "a", 0))
	assert.Equal(2, dict.AVUtil_av_dict_count())
	dict.AVUtil_av_dict_free()
}

func Test_avutil_005(t *testing.T) {
	assert := assert.New(t)
	dict := ffmpeg.AVUtil_av_dict_new()
	assert.NoError(dict.AVUtil_av_dict_set("a", "b", 0))
	assert.NoError(dict.AVUtil_av_dict_set("b", "a", 0))
	assert.Equal(2, dict.AVUtil_av_dict_count())
	assert.EqualValues([]string{"a", "b"}, dict.AVUtil_av_dict_keys())
	dict.AVUtil_av_dict_free()
}

func Test_avutil_006(t *testing.T) {
	assert := assert.New(t)
	fmts := []ffmpeg.AVSampleFormat{
		ffmpeg.AV_SAMPLE_FMT_NONE,
		ffmpeg.AV_SAMPLE_FMT_U8,
		ffmpeg.AV_SAMPLE_FMT_S16,
		ffmpeg.AV_SAMPLE_FMT_S32,
		ffmpeg.AV_SAMPLE_FMT_FLT,
		ffmpeg.AV_SAMPLE_FMT_DBL,
		ffmpeg.AV_SAMPLE_FMT_U8P,
		ffmpeg.AV_SAMPLE_FMT_S16P,
		ffmpeg.AV_SAMPLE_FMT_S32P,
		ffmpeg.AV_SAMPLE_FMT_FLTP,
		ffmpeg.AV_SAMPLE_FMT_DBLP,
		ffmpeg.AV_SAMPLE_FMT_S64,
		ffmpeg.AV_SAMPLE_FMT_S64P,
		ffmpeg.AV_SAMPLE_FMT_NB,
	}
	for _, fmt := range fmts {
		t.Log("fmt => ", fmt)
		if fmt != ffmpeg.AV_SAMPLE_FMT_NONE && fmt != ffmpeg.AV_SAMPLE_FMT_NB {
			str := ffmpeg.AVUtil_av_get_sample_fmt_name(fmt)
			assert.NotEqual("", str)
			t.Log("str => ", str)
			assert.Equal(fmt, ffmpeg.AVUtil_av_get_sample_fmt(str))
		}
	}
}

func Test_avutil_007(t *testing.T) {
	assert := assert.New(t)
	fmts := []ffmpeg.AVSampleFormat{
		ffmpeg.AV_SAMPLE_FMT_U8,
		ffmpeg.AV_SAMPLE_FMT_S16,
		ffmpeg.AV_SAMPLE_FMT_S32,
		ffmpeg.AV_SAMPLE_FMT_FLT,
		ffmpeg.AV_SAMPLE_FMT_DBL,
		ffmpeg.AV_SAMPLE_FMT_S64,
	}
	for _, packed := range fmts {
		t.Log("packed => ", packed)
		planar := ffmpeg.AVUtil_av_get_alt_sample_fmt(packed, true)
		assert.NotEqual(ffmpeg.AV_SAMPLE_FMT_NONE, planar)
		t.Log("  planar => ", planar)
		assert.Equal(packed, ffmpeg.AVUtil_av_get_alt_sample_fmt(planar, false))
		assert.Equal(true, ffmpeg.AVUtil_av_sample_fmt_is_planar(planar))
		assert.Equal(false, ffmpeg.AVUtil_av_sample_fmt_is_planar(packed))
	}
}
