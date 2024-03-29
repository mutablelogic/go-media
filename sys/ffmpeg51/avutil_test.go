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
	var dict *ffmpeg.AVDictionary
	var err error
	assert := assert.New(t)
	dict, err = ffmpeg.AVUtil_av_dict_set_ptr(dict, "a", "b", 0)
	assert.NoError(err)
	assert.NotNil(dict)
	dict, err = ffmpeg.AVUtil_av_dict_set_ptr(dict, "b", "b", 0)
	assert.NoError(err)
	assert.NotNil(dict)
	t.Log(dict)
	keys := ffmpeg.AVUtil_av_dict_keys(dict)
	assert.Equal(2, len(keys))
	entries := ffmpeg.AVUtil_av_dict_entries(dict)
	assert.Equal(2, len(entries))
	t.Log(keys)
	t.Log(entries)
	ffmpeg.AVUtil_av_dict_free_ptr(dict)
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

func Test_avutil_008(t *testing.T) {
	assert := assert.New(t)
	var iter uintptr
	for {
		layout := ffmpeg.AVUtil_av_channel_layout_standard(&iter)
		if layout == nil {
			break
		}
		str, err := ffmpeg.AVUtil_av_channel_layout_describe(layout)
		assert.NoError(err)
		assert.NotEqual("", str)
		t.Log("layout => ", str)
	}
}

func Test_avutil_009(t *testing.T) {
	assert := assert.New(t)
	var ch_layout ffmpeg.AVChannelLayout
	for ch := 1; ch <= 8; ch++ {
		ffmpeg.AVUtil_av_channel_layout_default(&ch_layout, ch)
		str, err := ffmpeg.AVUtil_av_channel_layout_describe(&ch_layout)
		assert.NoError(err)
		assert.NotEqual("", str)
		assert.Equal(true, ffmpeg.AVUtil_av_channel_layout_check(&ch_layout))
		assert.NoError(ffmpeg.AVUtil_av_channel_layout_from_string(&ch_layout, str))
		t.Log("ch=>", ch, " layout=>", str)
		for i := 0; i < ffmpeg.AVUtil_av_get_channel_layout_nb_channels(&ch_layout); i++ {
			ch := ffmpeg.AVUtil_av_channel_layout_channel_from_index(&ch_layout, i)
			name, err := ffmpeg.AVUtil_av_channel_name(ch)
			assert.NoError(err)
			assert.NotEqual("", name)
			description, err := ffmpeg.AVUtil_av_channel_description(ch)
			assert.NoError(err)
			assert.NotEqual("", description)
			t.Log("  ", i, "=> ch:", ch, "name:", name, "description:", description)
		}
	}
}

func Test_avutil_010(t *testing.T) {
	t.Log("avutil_version=", ffmpeg.AVUtil_version())
}

func Test_avutil_011(t *testing.T) {
	t.Log("avutil_configuration=", ffmpeg.AVUtil_configuration())
}

func Test_avutil_012(t *testing.T) {
	t.Log("avutil_license=", ffmpeg.AVUtil_license())
}
