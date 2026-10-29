package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	// Packages
	reader "github.com/mutablelogic/go-media/reader"
	segmenter "github.com/mutablelogic/go-media/segmenter"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// SegmentOpts are the options for segmenting audio
type SegmentOpts struct {
	Rate        int             `name:"rate" default:"16000" help:"Sample rate of the segments, in Hz."`
	Mode        string          `name:"mode" enum:"none,silence,quality,lowbitrate,aggressive,veryaggressive" default:"aggressive" help:"How silence is detected: none, silence (RMS energy), or a voice activity detection mode (quality, lowbitrate, aggressive, veryaggressive), which also classifies segments."`
	Silence     time.Duration   `name:"silence" help:"Duration of silence to cut segments at (default 500ms)."`
	Threshold   float64         `name:"threshold" help:"RMS energy (0.0-1.0) below which audio is silence (default 0.01 with --mode silence, none otherwise)."`
	Segments    []time.Duration `name:"segments" sep:"," help:"Segment duration or, with more than one, the times to cut segments near, e.g. 3m10s,7m45s."`
	Stream      *int            `name:"stream" help:"Index of the audio stream to decode (default the first)."`
	InputFormat string          `name:"input-format" help:"Input format name, for raw audio (e.g. s16le)."`
	InputOpts   []string        `name:"input-opts" help:"Input format options, for raw audio (e.g. sample_rate=16000,channels=1)."`
}

// feature is a stat which is written for each segment
type feature struct {
	name  string
	value func(segmenter.Stats) float64
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var modes = map[string]segmenter.Mode{
	"none":           segmenter.ModeNone,
	"silence":        segmenter.ModeSilence,
	"quality":        segmenter.ModeVADQuality,
	"lowbitrate":     segmenter.ModeVADLowBitrate,
	"aggressive":     segmenter.ModeVADAggressive,
	"veryaggressive": segmenter.ModeVADVeryAggressive,
}

var features = []feature{
	{"speech", func(s segmenter.Stats) float64 { return s.Speech }},
	{"rms", func(s segmenter.Stats) float64 { return s.RMS }},
	{"peak", func(s segmenter.Stats) float64 { return s.Peak }},
	{"flatness", func(s segmenter.Stats) float64 { return s.Flatness }},
	{"zcr", func(s segmenter.Stats) float64 { return s.ZCR }},
	{"zcr_variation", func(s segmenter.Stats) float64 { return s.ZCRVariation }},
	{"low_energy", func(s segmenter.Stats) float64 { return s.LowEnergy }},
	{"modulation", func(s segmenter.Stats) float64 { return s.Modulation }},
	{"centroid", func(s segmenter.Stats) float64 { return s.Centroid }},
	{"flux", func(s segmenter.Stats) float64 { return s.Flux }},
	{"flux_variation", func(s segmenter.Stats) float64 { return s.FluxVariation }},
	{"rhythm", func(s segmenter.Stats) float64 { return s.Rhythm }},
	{"harmonicity", func(s segmenter.Stats) float64 { return s.Harmonicity }},
	{"voiced", func(s segmenter.Stats) float64 { return s.Voiced }},
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// options returns the segmenter options
func (o SegmentOpts) options() ([]segmenter.Opt, error) {
	result, err := o.segmentOptions()
	if err != nil {
		return nil, err
	}
	if o.Stream != nil {
		result = append(result, segmenter.WithStream(*o.Stream))
	}
	if o.InputFormat != "" || len(o.InputOpts) > 0 {
		result = append(result, segmenter.WithReaderOpt(reader.WithInput(o.InputFormat, o.InputOpts...)))
	}
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// segmentOptions returns the segmenter options for how segments are cut,
// without those for reading the audio
func (o SegmentOpts) segmentOptions() ([]segmenter.Opt, error) {
	mode, exists := modes[o.Mode]
	if !exists {
		return nil, fmt.Errorf("invalid mode %q", o.Mode)
	}
	result := []segmenter.Opt{segmenter.WithMode(mode, o.Silence, o.Threshold)}
	if len(o.Segments) > 0 {
		result = append(result, segmenter.WithSegments(o.Segments...))
	}
	return result, nil
}

// header returns the CSV header for segments, after the leading columns
func header(leading ...string) []string {
	result := append(leading, "start", "duration", "class")
	for _, f := range features {
		result = append(result, f.name)
	}
	return result
}

// record returns the CSV record for a segment, after the leading columns
func record(s segmenter.Segment, leading ...string) []string {
	result := append(leading, seconds(s.Start), seconds(s.Duration), className(s.Class))
	for _, f := range features {
		result = append(result, strconv.FormatFloat(f.value(s.Stats), 'f', 4, 64))
	}
	return result
}

func seconds(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 3, 64)
}

func className(c segmenter.Class) string {
	return strings.ToLower(strings.TrimPrefix(c.String(), "Class"))
}
