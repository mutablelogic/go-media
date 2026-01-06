package ffmpeg

import (
	"errors"
	"fmt"
	"syscall"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Filter applies a filter graph to frames for a single stream.
// Supports single input → filter chain → single output.
type Filter struct {
	t          media.Type
	audio      *audioFilter
	video      *videoFilter
	filterSpec string
}

type audioFilter struct {
	graph   *ff.AVFilterGraph
	src     *ff.AVFilterContext
	sink    *ff.AVFilterContext
	srcPar  *Par
	destPar *Par
}

type videoFilter struct {
	graph   *ff.AVFilterGraph
	src     *ff.AVFilterContext
	sink    *ff.AVFilterContext
	srcPar  *Par
	destPar *Par
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new filter which will apply the given filter specification to frames.
// filterSpec is an FFmpeg filter string (e.g., "scale=1280:720", "volume=0.5").
// srcPar defines the input frame format, destPar defines the desired output format.
func NewFilter(filterSpec string, srcPar, destPar *Par) (*Filter, error) {
	if srcPar == nil {
		return nil, errors.New("srcPar is nil")
	}
	if destPar == nil {
		return nil, errors.New("destPar is nil")
	}
	if srcPar.Type() != destPar.Type() {
		return nil, errors.New("source and destination types must match")
	}

	f := &Filter{
		t:          srcPar.Type(),
		filterSpec: filterSpec,
	}

	switch f.t {
	case media.AUDIO:
		a, err := newAudioFilter(filterSpec, srcPar, destPar)
		if err != nil {
			return nil, err
		}
		f.audio = a
	case media.VIDEO:
		v, err := newVideoFilter(filterSpec, srcPar, destPar)
		if err != nil {
			return nil, err
		}
		f.video = v
	default:
		return nil, fmt.Errorf("unsupported type: %v", srcPar.Type())
	}

	return f, nil
}

// Release resources
func (f *Filter) Close() error {
	var err error
	if f.audio != nil {
		err = errors.Join(err, f.audio.Close())
	}
	if f.video != nil {
		err = errors.Join(err, f.video.Close())
	}
	f.audio = nil
	f.video = nil
	return err
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Process applies the filter to one frame and emits zero or more frames via fn.
// Pass src==nil to flush. fn may receive nil to signal end-of-batch.
func (f *Filter) Process(src *Frame, fn func(*Frame) error) error {
	if fn == nil {
		return media.ErrBadParameter.With("nil callback")
	}
	switch f.t {
	case media.AUDIO:
		return f.audio.process(src, fn)
	case media.VIDEO:
		return f.video.process(src, fn)
	default:
		return fmt.Errorf("unsupported type: %v", f.t)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - AUDIO

func newAudioFilter(filterSpec string, srcPar, destPar *Par) (*audioFilter, error) {
	if srcPar.CodecType() != ff.AVMEDIA_TYPE_AUDIO {
		return nil, errors.New("invalid source codec type")
	}
	if destPar.CodecType() != ff.AVMEDIA_TYPE_AUDIO {
		return nil, errors.New("invalid destination codec type")
	}

	// Allocate filter graph
	graph := ff.AVFilterGraph_alloc()
	if graph == nil {
		return nil, errors.New("failed to allocate filter graph")
	}

	// Get buffer and buffersink filters
	abuffer := ff.AVFilter_get_by_name("abuffer")
	if abuffer == nil {
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("abuffer filter not found")
	}
	abuffersink := ff.AVFilter_get_by_name("abuffersink")
	if abuffersink == nil {
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("abuffersink filter not found")
	}

	// Create buffer source arguments
	ch := srcPar.ChannelLayout()
	chLayout, _ := ff.AVUtil_channel_layout_describe(&ch)
	// Get timebase - use a default if not set
	tb := ff.AVUtil_rational(1, srcPar.SampleRate())
	srcArgs := fmt.Sprintf("sample_rate=%d:sample_fmt=%s:channel_layout=%s:time_base=%d/%d",
		srcPar.SampleRate(),
		ff.AVUtil_get_sample_fmt_name(srcPar.SampleFormat()),
		chLayout,
		tb.Num(),
		tb.Den(),
	)

	// Create source filter context
	src, err := ff.AVFilterGraph_create_filter(graph, abuffer, "src", srcArgs)
	if err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to create abuffer: %w", err)
	}

	// Create sink filter context
	sink, err := ff.AVFilterGraph_create_filter(graph, abuffersink, "sink", "")
	if err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to create abuffersink: %w", err)
	}

	// Parse and connect the filter graph
	fullSpec := fmt.Sprintf("[src]%s[sink]", filterSpec)
	inputs, outputs, err := ff.AVFilterGraph_parse(graph, fullSpec)
	if err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to parse filter graph: %w", err)
	}

	// Link the inputs and outputs
	if len(inputs) != 1 || inputs[0].Name() != "src" {
		ff.AVFilterInOut_list_free(inputs)
		ff.AVFilterInOut_list_free(outputs)
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("filter graph must have exactly one input named 'src'")
	}
	if len(outputs) != 1 || outputs[0].Name() != "sink" {
		ff.AVFilterInOut_list_free(inputs)
		ff.AVFilterInOut_list_free(outputs)
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("filter graph must have exactly one output named 'sink'")
	}

	// Free the inputs and outputs (they're already linked by parse)
	ff.AVFilterInOut_list_free(inputs)
	ff.AVFilterInOut_list_free(outputs)

	// Configure the filter graph
	if err := ff.AVFilterGraph_config(graph); err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to configure filter graph: %w", err)
	}

	return &audioFilter{
		graph:   graph,
		src:     src,
		sink:    sink,
		srcPar:  srcPar,
		destPar: destPar,
	}, nil
}

func (f *audioFilter) Close() error {
	if f.graph != nil {
		ff.AVFilterGraph_free(f.graph)
	}
	f.graph = nil
	f.src = nil
	f.sink = nil
	return nil
}

func (f *audioFilter) process(src *Frame, fn func(*Frame) error) error {
	if src != nil && src.Type() != media.AUDIO {
		return errors.New("frame type mismatch")
	}

	// Push frame into buffer source
	if src != nil {
		if err := ff.AVBufferSrc_add_frame_flags(f.src, (*ff.AVFrame)(src), ff.AV_BUFFERSRC_FLAG_KEEP_REF); err != nil {
			return fmt.Errorf("AVBufferSrc_add_frame: %w", err)
		}
	} else {
		// Flush by sending nil
		if err := ff.AVBufferSrc_add_frame_flags(f.src, nil, 0); err != nil {
			return fmt.Errorf("AVBufferSrc_add_frame (flush): %w", err)
		}
	}

	// Pull frames from buffersink
	for {
		frame := ff.AVUtil_frame_alloc()
		if frame == nil {
			return errors.New("failed to allocate frame")
		}

		err := ff.AVBufferSink_get_frame(f.sink, frame)
		if err != nil {
			ff.AVUtil_frame_free(frame)
			avErr, ok := err.(ff.AVError)
			if ok && (avErr == ff.AVERROR_EOF || avErr.IsErrno(syscall.EAGAIN)) {
				// No more frames available
				return nil
			}
			return fmt.Errorf("AVBufferSink_get_frame: %w", err)
		}

		// Call the callback with the filtered frame
		if err := fn((*Frame)(frame)); err != nil {
			ff.AVUtil_frame_free(frame)
			return err
		}

		ff.AVUtil_frame_free(frame)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - VIDEO

func newVideoFilter(filterSpec string, srcPar, destPar *Par) (*videoFilter, error) {
	if srcPar.CodecType() != ff.AVMEDIA_TYPE_VIDEO {
		return nil, errors.New("invalid source codec type")
	}
	if destPar.CodecType() != ff.AVMEDIA_TYPE_VIDEO {
		return nil, errors.New("invalid destination codec type")
	}

	// Allocate filter graph
	graph := ff.AVFilterGraph_alloc()
	if graph == nil {
		return nil, errors.New("failed to allocate filter graph")
	}

	// Get buffer and buffersink filters
	buffer := ff.AVFilter_get_by_name("buffer")
	if buffer == nil {
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("buffer filter not found")
	}
	buffersink := ff.AVFilter_get_by_name("buffersink")
	if buffersink == nil {
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("buffersink filter not found")
	}

	// Create buffer source arguments
	// Get timebase - calculate from frame rate if available
	tb := ff.AVUtil_rational(1, 25) // Default 25fps
	if fr := srcPar.FrameRate(); fr > 0 {
		tb = ff.AVUtil_rational_invert(ff.AVUtil_rational_d2q(fr, 1<<24))
	}
	srcArgs := fmt.Sprintf("video_size=%dx%d:pix_fmt=%d:time_base=%d/%d:pixel_aspect=%d/%d",
		srcPar.Width(),
		srcPar.Height(),
		srcPar.PixelFormat(),
		tb.Num(),
		tb.Den(),
		srcPar.SampleAspectRatio().Num(),
		srcPar.SampleAspectRatio().Den(),
	)

	// Create source filter context
	src, err := ff.AVFilterGraph_create_filter(graph, buffer, "src", srcArgs)
	if err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to create buffer: %w", err)
	}

	// Create sink filter context
	sink, err := ff.AVFilterGraph_create_filter(graph, buffersink, "sink", "")
	if err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to create buffersink: %w", err)
	}

	// Parse and connect the filter graph
	fullSpec := fmt.Sprintf("[src]%s[sink]", filterSpec)
	inputs, outputs, err := ff.AVFilterGraph_parse(graph, fullSpec)
	if err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to parse filter graph: %w", err)
	}

	// Link the inputs and outputs
	if len(inputs) != 1 || inputs[0].Name() != "src" {
		ff.AVFilterInOut_list_free(inputs)
		ff.AVFilterInOut_list_free(outputs)
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("filter graph must have exactly one input named 'src'")
	}
	if len(outputs) != 1 || outputs[0].Name() != "sink" {
		ff.AVFilterInOut_list_free(inputs)
		ff.AVFilterInOut_list_free(outputs)
		ff.AVFilterGraph_free(graph)
		return nil, errors.New("filter graph must have exactly one output named 'sink'")
	}

	// Explicitly link the buffer source and sink to the parsed filter graph.
	// Connect: src (out pad 0) -> first filter in the graph.
	if err := ff.AVFilterContext_link(
		src,
		0,
		inputs[0].FilterCtx(),
		inputs[0].PadIdx(),
	); err != nil {
		ff.AVFilterInOut_list_free(inputs)
		ff.AVFilterInOut_list_free(outputs)
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to link source filter: %w", err)
	}

	// Connect: last filter in the graph -> sink (in pad 0).
	if err := ff.AVFilterContext_link(
		outputs[0].FilterCtx(),
		outputs[0].PadIdx(),
		sink,
		0,
	); err != nil {
		ff.AVFilterInOut_list_free(inputs)
		ff.AVFilterInOut_list_free(outputs)
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to link sink filter: %w", err)
	}

	// Free the inputs and outputs now that they have been linked
	ff.AVFilterInOut_list_free(inputs)
	ff.AVFilterInOut_list_free(outputs)

	// Configure the filter graph
	if err := ff.AVFilterGraph_config(graph); err != nil {
		ff.AVFilterGraph_free(graph)
		return nil, fmt.Errorf("failed to configure filter graph: %w", err)
	}

	return &videoFilter{
		graph:   graph,
		src:     src,
		sink:    sink,
		srcPar:  srcPar,
		destPar: destPar,
	}, nil
}

func (f *videoFilter) Close() error {
	if f.graph != nil {
		ff.AVFilterGraph_free(f.graph)
	}
	f.graph = nil
	f.src = nil
	f.sink = nil
	return nil
}

func (f *videoFilter) process(src *Frame, fn func(*Frame) error) error {
	if src != nil && src.Type() != media.VIDEO {
		return errors.New("frame type mismatch")
	}

	// Push frame into buffer source
	if src != nil {
		if err := ff.AVBufferSrc_add_frame_flags(f.src, (*ff.AVFrame)(src), ff.AV_BUFFERSRC_FLAG_KEEP_REF); err != nil {
			return fmt.Errorf("AVBufferSrc_add_frame: %w", err)
		}
	} else {
		// Flush by sending nil
		if err := ff.AVBufferSrc_add_frame_flags(f.src, nil, 0); err != nil {
			return fmt.Errorf("AVBufferSrc_add_frame (flush): %w", err)
		}
	}

	// Pull frames from buffersink
	for {
		frame := ff.AVUtil_frame_alloc()
		if frame == nil {
			return errors.New("failed to allocate frame")
		}

		err := ff.AVBufferSink_get_frame(f.sink, frame)
		if err != nil {
			ff.AVUtil_frame_free(frame)
			avErr, ok := err.(ff.AVError)
			if ok && (avErr == ff.AVERROR_EOF || avErr.IsErrno(syscall.EAGAIN)) {
				// No more frames available
				return nil
			}
			return fmt.Errorf("AVBufferSink_get_frame: %w", err)
		}

		// Call the callback with the filtered frame
		if err := fn((*Frame)(frame)); err != nil {
			ff.AVUtil_frame_free(frame)
			return err
		}

		ff.AVUtil_frame_free(frame)
	}
}
