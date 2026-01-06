package ffmpeg_test

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

func Test_avfilter_buffersrc_000(t *testing.T) {
	assert := assert.New(t)

	// Create a simple filter graph with buffer source and buffersink
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	// Get buffer and buffersink filters
	buffer := ff.AVFilter_get_by_name("buffer")
	assert.NotNil(buffer)
	buffersink := ff.AVFilter_get_by_name("buffersink")
	assert.NotNil(buffersink)

	// Create source and sink contexts
	src, err := ff.AVFilterGraph_create_filter(graph, buffer, "src", "video_size=320x240:pix_fmt=0:time_base=1/25:pixel_aspect=1/1")
	assert.NoError(err)
	assert.NotNil(src)

	sink, err := ff.AVFilterGraph_create_filter(graph, buffersink, "sink", "")
	assert.NoError(err)
	assert.NotNil(sink)

	t.Log("buffer source=", src)
	t.Log("buffer sink=", sink)
}

func Test_avfilter_buffersrc_001(t *testing.T) {
	assert := assert.New(t)

	// Test complete video filtering pipeline with scale
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	buffer := ff.AVFilter_get_by_name("buffer")
	buffersink := ff.AVFilter_get_by_name("buffersink")
	assert.NotNil(buffer)
	assert.NotNil(buffersink)

	// Create source and sink
	src, err := ff.AVFilterGraph_create_filter(graph, buffer, "src", "video_size=320x240:pix_fmt=0:time_base=1/25:pixel_aspect=1/1")
	assert.NoError(err)

	sink, err := ff.AVFilterGraph_create_filter(graph, buffersink, "sink", "")
	assert.NoError(err)

	// Create scale filter
	scale := ff.AVFilter_get_by_name("scale")
	assert.NotNil(scale)
	scaleCtx, err := ff.AVFilterGraph_create_filter(graph, scale, "scale", "640:480")
	assert.NoError(err)

	// Connect: src -> scale -> sink
	err = ff.AVFilterContext_link(src, 0, scaleCtx, 0)
	if err != nil {
		t.Fatalf("Link src->scale failed: %v", err)
	}
	err = ff.AVFilterContext_link(scaleCtx, 0, sink, 0)
	if err != nil {
		t.Fatalf("Link scale->sink failed: %v", err)
	}

	// Configure
	err = ff.AVFilterGraph_config(graph)
	assert.NoError(err)

	// Create and add a frame
	frame := ff.AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer ff.AVUtil_frame_free(frame)

	frame.SetWidth(320)
	frame.SetHeight(240)
	frame.SetPixFmt(ff.AV_PIX_FMT_YUV420P)

	err = ff.AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Add frame to buffer source
	err = ff.AVBufferSrc_add_frame_flags(src, frame, ff.AV_BUFFERSRC_FLAG_KEEP_REF)
	assert.NoError(err)

	// Get filtered frame from sink
	outFrame := ff.AVUtil_frame_alloc()
	assert.NotNil(outFrame)
	defer ff.AVUtil_frame_free(outFrame)

	err = ff.AVBufferSink_get_frame(sink, outFrame)
	assert.NoError(err)

	// Verify output frame dimensions
	assert.Equal(640, int(outFrame.Width()))
	assert.Equal(480, int(outFrame.Height()))

	t.Logf("Filtered frame: %dx%d", outFrame.Width(), outFrame.Height())
}

func Test_avfilter_buffersrc_002(t *testing.T) {
	assert := assert.New(t)

	// Test complete audio filtering pipeline
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	abuffer := ff.AVFilter_get_by_name("abuffer")
	assert.NotNil(abuffer)
	abuffersink := ff.AVFilter_get_by_name("abuffersink")
	assert.NotNil(abuffersink)

	// Create audio source and sink
	src, err := ff.AVFilterGraph_create_filter(graph, abuffer, "src", "sample_rate=44100:sample_fmt=fltp:channel_layout=stereo:time_base=1/44100")
	assert.NoError(err)
	assert.NotNil(src)

	sink, err := ff.AVFilterGraph_create_filter(graph, abuffersink, "sink", "")
	assert.NoError(err)
	assert.NotNil(sink)

	// Connect src to sink directly
	err = ff.AVFilterContext_link(src, 0, sink, 0)
	assert.NoError(err)

	// Configure the graph
	err = ff.AVFilterGraph_config(graph)
	assert.NoError(err)

	// Test audio-specific getter
	sampleRate := ff.AVBufferSink_get_sample_rate(sink)
	assert.Equal(44100, sampleRate)

	t.Logf("Audio sample rate: %d", sampleRate)
}
