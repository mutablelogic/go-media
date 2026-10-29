package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	segmenter "github.com/mutablelogic/go-media/segmenter"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// OutputOpts are the options for what's written for each segment
type OutputOpts struct {
	Out    string `arg:"" optional:"" name:"out" type:"path" help:"Folder to write each segment to, as a file."`
	Format string `name:"format" enum:"wav,mp3,aac" default:"aac" help:"Format of the segment files: wav, mp3 or aac."`
	CSV    string `name:"csv" type:"path" help:"Write the class and stats of each segment to this CSV file."`
}

// output writes the file and the CSV record for each segment
type output struct {
	OutputOpts
	sampleRate int
	file       *os.File
	csv        *csv.Writer
	n          int // number of segments written
}

// encoding is how segment files are encoded
type encoding struct {
	container, codec, format, ext string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var encodings = map[string]encoding{
	"wav": {"wav", "pcm_s16le", "s16", "wav"},
	"mp3": {"mp3", "libmp3lame", "fltp", "mp3"},
	"aac": {"adts", "aac", "fltp", "aac"},
}

// rowFormat is the format of each row printed for a segment, and fileFormat
// is the format of the file it's written to, if it is
const (
	rowFormat  = "%-7v  %-10v  %-10v  %-10v  %-7v  %6v  %6v"
	fileFormat = "  %v"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// open creates the folder and CSV file for segments at a sample rate
func (o OutputOpts) open(sampleRate int) (*output, error) {
	result := &output{OutputOpts: o, sampleRate: sampleRate}
	if o.Out != "" {
		if err := os.MkdirAll(o.Out, 0o755); err != nil {
			return nil, err
		}
	}
	if o.CSV != "" {
		f, err := os.Create(o.CSV)
		if err != nil {
			return nil, err
		}
		result.file, result.csv = f, csv.NewWriter(f)
		if err := result.csv.Write(header("segment")); err != nil {
			return nil, errors.Join(err, f.Close())
		}
	}
	return result, nil
}

// Close writes and closes the CSV file
func (o *output) Close() error {
	if o.file == nil {
		return nil
	}
	o.csv.Flush()
	return errors.Join(o.csv.Error(), o.file.Close())
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// header returns the row printed before the segments
func (o *output) header() string {
	return o.row("FILE", "SEGMENT", "START", "END", "DURATION", "CLASS", "SPEECH", "RMS")
}

// row returns a row to print, with the file if segments are written to files
func (o *output) row(file string, values ...any) string {
	result := fmt.Sprintf(rowFormat, values...)
	if o.Out != "" {
		result += fmt.Sprintf(fileFormat, file)
	}
	return result + "\n"
}

// write writes the file and CSV record for a segment, and returns the row to
// print for it
func (o *output) write(s *segmenter.Segment) (string, error) {
	o.n++
	var path string
	if o.Out != "" {
		path = filepath.Join(o.Out, fmt.Sprintf("%03d-%s.%s", o.n, className(s.Class), encodings[o.Format].ext))
		if err := encode(path, encodings[o.Format], o.sampleRate, s.Samples); err != nil {
			return "", err
		}
	}
	if o.csv != nil {
		if err := o.csv.Write(record(*s, fmt.Sprint(o.n))); err != nil {
			return "", err
		}
	}
	return o.row(path, o.n, round(s.Start), round(s.Start+s.Duration), round(s.Duration), className(s.Class), fmt.Sprintf("%.2f", s.Stats.Speech), fmt.Sprintf("%.3f", s.Stats.RMS)), nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// round returns a duration rounded to the nearest 10ms
func round(d time.Duration) time.Duration {
	return d.Round(10 * time.Millisecond)
}

// encode writes mono samples at a sample rate to a file
func encode(path string, e encoding, sampleRate int, samples []float32) (err error) {
	output := profile.OutputWithName(e.container)
	if output == nil {
		return fmt.Errorf("format %q is not supported", e.container)
	}
	p, err := profile.NewAudioProfile(e.codec)
	if err != nil {
		return err
	}
	if err := p.Set(profile.OptionSampleRate, uint64(sampleRate)); err != nil {
		return err
	}
	if err := p.Set(profile.OptionSampleFormat, e.format); err != nil {
		return err
	}
	if err := p.Set(profile.OptionChannelLayout, "mono"); err != nil {
		return err
	}

	w, err := writer.Create(&url.URL{Path: path}, output, writer.WithProfile(0, p))
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, w.Close())
	}()

	// Encode the samples as one frame, which the writer resamples into
	// frames of the size the codec needs
	f, err := frame.NewAudioFrame(0)
	if err != nil {
		return err
	}
	defer f.Close()
	f.SetSampleFormat(ff.AV_SAMPLE_FMT_FLT)
	f.SetSampleRate(sampleRate)
	var layout ff.AVChannelLayout
	ff.AVUtil_channel_layout_default(&layout, 1)
	if err := f.SetChannelLayout(layout); err != nil {
		return err
	}
	f.SetNumSamples(len(samples))
	if err := f.AllocateBuffers(); err != nil {
		return err
	}
	copy(f.Float32(0), samples)
	f.SetPts(0)

	if err := w.Encode(f); err != nil {
		return err
	}
	return w.Flush(0)
}
