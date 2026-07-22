package writer_test

import (
	"testing"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

func silentFrame(t *testing.T, streamID, numSamples int) *frame.Frame {
	t.Helper()

	frame, err := frame.NewFrame(streamID)
	if err != nil {
		t.Fatalf("NewFrame: %v", err)
	}

	frame.SetSampleFormat(ff.AVUtil_get_sample_fmt("fltp"))
	frame.SetSampleRate(44100)

	var ch ff.AVChannelLayout
	if err := ff.AVUtil_channel_layout_from_string(&ch, "stereo"); err != nil {
		t.Fatalf("AVUtil_channel_layout_from_string: %v", err)
	}
	if err := frame.SetChannelLayout(ch); err != nil {
		t.Fatalf("SetChannelLayout: %v", err)
	}
	frame.SetNumSamples(numSamples)

	if err := frame.AllocateBuffers(); err != nil {
		t.Fatalf("AllocateBuffers: %v", err)
	}

	for plane := 0; plane < frame.NumChannels(); plane++ {
		samples := frame.Float32(plane)
		for i := range samples {
			samples[i] = 0
		}
	}

	return frame
}

func newTestEncoder(t *testing.T, fn writer.PacketFn) *writer.Encoder {
	t.Helper()
	if fn == nil {
		fn = func(*ff.AVPacket) error { return nil }
	}
	enc, err := writer.NewEncoder(fn)
	if err != nil {
		t.Fatalf("NewEncoder: %v", err)
	}
	return enc
}

func TestEncoder(t *testing.T) {
	var packets int
	enc := newTestEncoder(t, func(pkt *ff.AVPacket) error {
		if pkt != nil {
			packets++
		}
		return nil
	})
	defer enc.Close()

	if err := enc.Add(0, audioStream(t)); err != nil {
		t.Fatalf("Add: %v", err)
	}

	numSamples := enc.FrameSize(0)
	if numSamples == 0 {
		numSamples = 1024
	}

	for i := 0; i < 5; i++ {
		frame := silentFrame(t, 0, numSamples)
		frame.SetPts(int64(i * numSamples))

		if err := enc.Encode(frame); err != nil {
			frame.Close()
			t.Fatalf("Encode: %v", err)
		}
		frame.Close()
	}

	if err := enc.Flush(0); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if packets == 0 {
		t.Fatal("expected at least one packet")
	}
}

func TestNewEncoder_NilFn(t *testing.T) {
	if _, err := writer.NewEncoder(nil); err == nil {
		t.Fatal("NewEncoder: expected error for nil callback function")
	}
}

func TestEncoder_Add_NilProfile(t *testing.T) {
	enc := newTestEncoder(t, nil)
	defer enc.Close()

	if err := enc.Add(0, nil); err == nil {
		t.Fatal("Add: expected error for nil profile")
	}
}

func TestEncoder_Add_Duplicate(t *testing.T) {
	enc := newTestEncoder(t, nil)
	defer enc.Close()

	if err := enc.Add(0, audioStream(t)); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := enc.Add(0, audioStream(t)); err == nil {
		t.Fatal("Add: expected error for duplicate stream id")
	}
}

func TestEncoder_Encode_UnknownStream(t *testing.T) {
	enc := newTestEncoder(t, nil)
	defer enc.Close()

	frame := silentFrame(t, 99, 1024)
	defer frame.Close()

	if err := enc.Encode(frame); err == nil {
		t.Fatal("Encode: expected error for unregistered stream")
	}
}

func TestEncoder_Close_Idempotent(t *testing.T) {
	enc := newTestEncoder(t, nil)
	if err := enc.Add(0, audioStream(t)); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("Close (first): %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("Close (second): %v", err)
	}
}
