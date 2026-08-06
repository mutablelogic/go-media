package encoder

import (
	"errors"
	"net/url"
	"os"
	"sort"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	task "github.com/mutablelogic/go-media/task/schema"
	writer "github.com/mutablelogic/go-media/writer"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Validate checks that req is well-formed: a non-nil Reader, a resolved
// Output, and at least one of Audio/Video/Subtitle set. Called by
// Manager.Add before the task is even registered, and again by Run itself
// as a safety net for any caller that builds a task without going through
// Add (e.g. Run called directly in a test).
func (req *EncodeRequest) Validate() error {
	if req.Reader == nil {
		return gomedia.ErrBadParameter.With("nil reader")
	}
	if req.Output == nil || req.Output.Context() == nil {
		return gomedia.ErrBadParameter.With("missing output format")
	}
	if req.Audio == nil && req.Video == nil && req.Subtitle == nil {
		return gomedia.ErrBadParameter.With("at least one of audio, video or subtitle profile must be set")
	}
	return nil
}

func (req *EncodeRequest) Run(ctx task.Context) (err error) {
	if err := req.Validate(); err != nil {
		return err
	}

	// Open the input
	seeker, err := task.NewReadSeeker(req.Reader)
	if err != nil {
		return err
	}
	rd, err := reader.NewReader(seeker)
	if err != nil {
		return err
	}
	defer rd.Close()
	streams := rd.Streams()

	// Match each requested profile against the first input stream of its
	// type. The writer needs its streams' ids to be exactly {0, ..., n-1} -
	// AVFormat_new_stream assigns each output AVStream's physical index by
	// creation order, regardless of the id WithProfile was given, so a
	// matched input stream's own (possibly non-zero, non-contiguous) index
	// can't be reused directly: e.g. selecting only an input's audio stream
	// (index 1, with video at 0) would tag every packet stream_index=1
	// against an output that only ever has a stream 0. remap gives each
	// matched input stream a dense output id instead, applied to each
	// decoded frame via Frame.SetStream before it reaches the writer.
	var opts []writer.Opt
	var inputIDs []int
	remap := make(map[int]int, 3)

	if req.Audio != nil {
		id, ok := firstStreamOfType(streams, ff.AVMEDIA_TYPE_AUDIO)
		if !ok {
			return gomedia.ErrBadParameter.With("no audio stream found in input")
		}
		inputIDs = append(inputIDs, id)
		remap[id] = len(remap)
		opts = append(opts, writer.WithProfile(remap[id], req.Audio))
	}
	if req.Video != nil {
		id, ok := firstStreamOfType(streams, ff.AVMEDIA_TYPE_VIDEO)
		if !ok {
			return gomedia.ErrBadParameter.With("no video stream found in input")
		}
		inputIDs = append(inputIDs, id)
		remap[id] = len(remap)
		opts = append(opts, writer.WithProfile(remap[id], req.Video))
	}
	if req.Subtitle != nil {
		id, ok := firstStreamOfType(streams, ff.AVMEDIA_TYPE_SUBTITLE)
		if !ok {
			return gomedia.ErrBadParameter.With("no subtitle stream found in input")
		}
		inputIDs = append(inputIDs, id)
		remap[id] = len(remap)
		opts = append(opts, writer.WithProfile(remap[id], req.Subtitle))
	}

	// Create the output - for now, a temporary file, until a real
	// upload/storage destination exists
	tmp, err := os.CreateTemp("", "gomedia-encode-*."+req.Output.Format)
	if err != nil {
		return err
	}
	path := tmp.Name()
	if err := tmp.Close(); err != nil {
		return err
	}

	w, err := writer.Create(&url.URL{Path: path}, req.Output, opts...)
	if err != nil {
		return err
	}

	// Decode the matched streams and hand every frame to the writer, first
	// relabelling it from its input stream index to the dense output id
	// remap assigned it - its per-stream Resampler already converts each
	// frame into the exact format its codec was opened with, including
	// re-chunking audio into exactly the number of samples the codec
	// expects. Progress is reported as the number of input frames decoded
	// so far - the total is left at 0 (unknown), since a reliable frame
	// count for the input isn't available up front.
	var frames uint64
	dec, err := reader.NewDecoder(func(f frame.Frame) error {
		frames++
		ctx.Progress(frames, 0)
		f.SetStream(remap[f.Stream()])
		return w.Encode(f)
	})
	if err != nil {
		return errors.Join(err, w.Close())
	}
	for _, id := range inputIDs {
		if err := dec.Add(id); err != nil {
			return errors.Join(err, w.Close())
		}
	}

	if err := rd.Decode(ctx, nil, dec); err != nil {
		return errors.Join(err, w.Close())
	}

	for _, id := range inputIDs {
		if err := w.Flush(remap[id]); err != nil {
			return errors.Join(err, w.Close())
		}
	}
	if err := w.Close(); err != nil {
		return err
	}

	// Set result
	ctx.Result(&EncodeResponse{Path: path})

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// firstStreamOfType returns the (real, demuxed) index of the first stream of
// the given media type, in stream order.
func firstStreamOfType(streams map[int]profile.Profile, want ff.AVMediaType) (int, bool) {
	ids := make([]int, 0, len(streams))
	for id := range streams {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		if ff.AVMediaType(streams[id].Type()) == want {
			return id, true
		}
	}
	return 0, false
}
