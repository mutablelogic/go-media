package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	// Packages
	reader "github.com/mutablelogic/go-media/reader"
	segmenter "github.com/mutablelogic/go-media/segmenter"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	server "github.com/mutablelogic/go-server"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type StreamCmd struct {
	Path      string        `arg:"" optional:"" name:"url" help:"URL of a stream, or a device such as device://avfoundation/:0 for the first microphone on macOS. Without it, the devices are listed."`
	Mode      string        `name:"mode" enum:"none,silence,quality,lowbitrate,aggressive,veryaggressive" default:"aggressive" help:"How silence is detected: none, silence (RMS energy), or a voice activity detection mode (quality, lowbitrate, aggressive, veryaggressive), which also classifies segments."`
	Silence   time.Duration `name:"silence" help:"Duration of silence to cut segments at (default 500ms)."`
	Threshold float64       `name:"threshold" default:"0.01" help:"RMS energy (0.0-1.0) below which audio is silence, so background noise isn't taken for sound."`
	Duration  time.Duration `name:"duration" help:"Stop after this duration."`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// streamSampleRate is the sample rate a stream is segmented at
const streamSampleRate = 16000

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run segments a stream or device until it ends, the duration has passed or
// it's interrupted, and shows each segment, without writing anything. The
// audio is decoded here, so its level can be shown when the output is a
// terminal, and passed on to the segmenter, which outputs the samples it has
// when the stream stops
func (cmd *StreamCmd) Run(ctx server.Cmd) error {
	if cmd.Path == "" {
		return listDevices(os.Stdout)
	}
	mode, exists := modes[cmd.Mode]
	if !exists {
		return fmt.Errorf("invalid mode %q", cmd.Mode)
	}
	out, err := OutputOpts{}.open(streamSampleRate)
	if err != nil {
		return err
	}

	// A device, such as a microphone, is "device://<format>/<address>"
	path, inputFormat := cmd.Path, ""
	if strings.HasPrefix(path, "device://") {
		format, address, _ := strings.Cut(strings.TrimPrefix(path, "device://"), "/")
		if format == "" || address == "" {
			return fmt.Errorf("invalid device %q, expected \"device://<format>/<address>\"", path)
		}
		path, inputFormat = address, format
	}

	// Show the level of the audio and the last segment, on a terminal
	var d *display
	print := func(line string, _ *segmenter.Segment) { fmt.Print(line) }
	var level func(level)
	if servercmd.TerminalWidth() > 0 {
		d = newDisplay(os.Stdout)
		defer d.Close()
		print, level = d.print, d.update
	}

	// Decode the audio, until the duration has passed or it's interrupted
	tapCtx := ctx.Context()
	if cmd.Duration > 0 {
		var cancel context.CancelFunc
		tapCtx, cancel = context.WithTimeout(tapCtx, cmd.Duration)
		defer cancel()
	}
	t, err := newTap(path, nil, streamSampleRate, inputFormat, nil, level)
	if err != nil {
		return err
	}
	defer t.Close()
	done := make(chan error, 1)
	go func() { done <- t.run(tapCtx) }()
	defer func() {
		t.stop()
		if err := <-done; err != nil && !errors.Is(err, io.ErrClosedPipe) {
			ctx.Logger().Error("stream", "error", err)
		}
	}()

	// Segment the decoded audio, until there's no more
	print(out.header(), nil)
	format, inputOpts := t.inputOpts()
	seg, err := segmenter.NewReader(t.Reader(), streamSampleRate,
		segmenter.WithMode(mode, cmd.Silence, cmd.Threshold),
		segmenter.WithReaderOpt(reader.WithInput(format, inputOpts...)),
	)
	if err != nil {
		return err
	}
	defer seg.Close()
	ctx.Logger().Debug("segmenting", "url", cmd.Path, "mode", cmd.Mode)

	return seg.Decode(context.Background(), func(s *segmenter.Segment) error {
		row, err := out.write(s)
		if err != nil {
			return err
		}
		print(row, s)
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// listDevices writes the URL of each audio input device, such as a
// microphone, which can be streamed
func listDevices(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintln(tw, "URL\tDEVICE\tDEFAULT")
	row := func(format, address, name string, isDefault bool) {
		var def string
		if isDefault {
			def = "yes"
		}
		fmt.Fprintf(tw, "device://%s/%s\t%s\t%s\n", format, address, name, def)
	}

	// Device formats are listed by separate audio and video iterators, but
	// may be the same format either way, such as avfoundation, which is
	// listed as a video device
	var n int
	seen := make(map[string]bool)
	add := func(d *ff.AVInputFormat, audio bool) {
		if seen[d.Name()] {
			return
		}
		seen[d.Name()] = true

		// avfoundation numbers video and audio devices separately, and an
		// audio device's address is ":<index>", with no video
		if d.Name() == "avfoundation" {
			for _, device := range ff.AVFoundationListDevices(d) {
				if device.MediaType == ff.AVMEDIA_TYPE_AUDIO {
					row(d.Name(), fmt.Sprintf(":%d", device.Index), device.Name, device.IsDefault)
					n++
				}
			}
			return
		}
		list, err := ff.AVDevice_list_input_sources(d, "", nil)
		if err != nil || list == nil {
			return
		}
		defer ff.AVDevice_free_list_devices(list)
		for i, device := range list.Devices() {
			if audio || slices.Contains(device.MediaTypes(), ff.AVMEDIA_TYPE_AUDIO) {
				row(d.Name(), device.Name(), device.Description(), list.Default() == i)
				n++
			}
		}
	}
	for d := ff.AVDevice_input_audio_device_first(); d != nil; d = ff.AVDevice_input_audio_device_next(d) {
		add(d, true)
	}
	for d := ff.AVDevice_input_video_device_first(); d != nil; d = ff.AVDevice_input_video_device_next(d) {
		add(d, false)
	}
	if n == 0 {
		return fmt.Errorf("no audio input devices")
	}
	return nil
}
