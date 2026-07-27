package task

import (
	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// audioFIFO buffers decoded audio samples and re-chunks them into
// fixed-size frames before handing them to encode - most frame-based audio
// codecs (e.g. AAC) require every frame but the last to carry exactly
// frameSize samples, which a decoder's own frame size rarely matches. A
// frameSize of 0 or less means the codec accepts any frame size, so samples
// are passed straight through unbuffered.
//
// audioFIFO does no resampling: it assumes every frame passed to write has
// the same sample format, rate and channel layout as the first.
type audioFIFO struct {
	streamID  int
	frameSize int
	encode    func(frame.Frame) error

	format         ff.AVSampleFormat
	sampleRate     int
	layout         ff.AVChannelLayout
	planar         bool
	numChannels    int
	bytesPerSample int
	planes         [][]byte // buffered raw sample bytes, one entry per plane
	pts            int64    // pts (in samples) of the next frame emitted
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// newAudioFIFO creates an audioFIFO for streamID that re-chunks buffered
// samples into frameSize-sized frames, passing each to encode.
func newAudioFIFO(streamID, frameSize int, encode func(frame.Frame) error) *audioFIFO {
	return &audioFIFO{streamID: streamID, frameSize: frameSize, encode: encode}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// write appends f's samples to the buffer, encoding as many full
// frameSize-sized frames as are now available. f is not retained or closed
// beyond this call - the caller remains responsible for it.
func (fifo *audioFIFO) write(f *frame.AudioFrame) error {
	if fifo.frameSize <= 0 {
		// No frame-size constraint - pass straight through unbuffered
		return fifo.encode(f)
	}

	if fifo.planes == nil {
		fifo.format = f.SampleFormat()
		fifo.sampleRate = f.SampleRate()
		fifo.layout = f.ChannelLayout()
		fifo.planar = ff.AVUtil_sample_fmt_is_planar(fifo.format)
		fifo.numChannels = f.NumChannels()
		fifo.bytesPerSample = ff.AVUtil_get_bytes_per_sample(fifo.format)

		numPlanes := 1
		if fifo.planar {
			numPlanes = fifo.numChannels
		}
		fifo.planes = make([][]byte, numPlanes)
	}

	for p := range fifo.planes {
		fifo.planes[p] = append(fifo.planes[p], f.Bytes(p)...)
	}

	return fifo.drain(false)
}

// flush encodes any buffered samples: first as full frameSize-sized frames,
// then a final, possibly short, frame for whatever remains. It's a no-op if
// write was never called, or the FIFO has no frame-size constraint (in
// which case there is never anything buffered to flush).
func (fifo *audioFIFO) flush() error {
	return fifo.drain(true)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// stride is the number of bytes one sample occupies across every channel a
// plane carries - just bytesPerSample for a planar format (one channel per
// plane), or bytesPerSample*numChannels for a packed one (every channel
// interleaved into plane 0).
func (fifo *audioFIFO) stride() int {
	if fifo.planar {
		return fifo.bytesPerSample
	}
	return fifo.bytesPerSample * fifo.numChannels
}

// available reports how many samples are currently buffered.
func (fifo *audioFIFO) available() int {
	if len(fifo.planes) == 0 || fifo.stride() == 0 {
		return 0
	}
	return len(fifo.planes[0]) / fifo.stride()
}

// drain emits full frameSize-sized frames for as long as enough samples are
// buffered, then - if final is true - one last frame for any remainder.
func (fifo *audioFIFO) drain(final bool) error {
	for fifo.available() >= fifo.frameSize {
		if err := fifo.take(fifo.frameSize); err != nil {
			return err
		}
	}
	if final && fifo.available() > 0 {
		return fifo.take(fifo.available())
	}
	return nil
}

// take builds a frame from the next n buffered samples, removes them from
// the buffer, and passes the frame to encode.
func (fifo *audioFIFO) take(n int) error {
	out, err := frame.NewAudioFrame(fifo.streamID)
	if err != nil {
		return err
	}
	defer out.Close()

	out.SetSampleFormat(fifo.format)
	out.SetSampleRate(fifo.sampleRate)
	if err := out.SetChannelLayout(fifo.layout); err != nil {
		return err
	}
	out.SetNumSamples(n)
	if err := out.AllocateBuffers(); err != nil {
		return err
	}

	need := n * fifo.stride()
	for p := range fifo.planes {
		copy(out.Bytes(p), fifo.planes[p][:need])
		fifo.planes[p] = fifo.planes[p][need:]
	}

	out.SetPts(fifo.pts)
	fifo.pts += int64(n)

	return fifo.encode(out)
}
