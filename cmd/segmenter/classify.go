package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	// Packages
	segmenter "github.com/mutablelogic/go-media/segmenter"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ClassifyCmd struct {
	SegmentOpts
	Folders []string `arg:"" name:"folder" help:"Folders of audio to classify, each named for the label of the audio in it, or label=folder."`
	Limit   int      `name:"limit" help:"Maximum number of files for each label, spread across the folder, or zero for all of them."`
	Jobs    int      `name:"jobs" short:"j" help:"Number of files to classify at once (default the number of CPUs)."`
	CSV     string   `name:"csv" type:"path" help:"Write the class and stats of each segment to this CSV file."`
}

type source struct {
	label string
	path  string
}

type row struct {
	source
	segmenter.Segment
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// defaultClassifySegment is the segment duration the classification
// thresholds were chosen with
const defaultClassifySegment = 2 * time.Second

var classes = []segmenter.Class{segmenter.ClassSilence, segmenter.ClassSpeech, segmenter.ClassMusic, segmenter.ClassOther, segmenter.ClassUnknown}

var extensions = []string{".wav", ".mp3", ".flac", ".ogg", ".m4a", ".aac"}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ClassifyCmd) Run(ctx server.Cmd) (err error) {
	if len(cmd.Segments) == 0 {
		cmd.Segments = []time.Duration{defaultClassifySegment}
	}
	opts, err := cmd.options()
	if err != nil {
		return err
	}

	// Find the files for each label
	var labels []string
	var sources []source
	for _, arg := range cmd.Folders {
		label, path, found := strings.Cut(arg, "=")
		if !found {
			label, path = filepath.Base(filepath.Clean(arg)), arg
		}
		files, err := find(path, cmd.Limit)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return fmt.Errorf("%s: no audio files", path)
		}
		ctx.Logger().Info("found audio", "label", label, "files", len(files))
		labels = append(labels, label)
		for _, file := range files {
			sources = append(sources, source{label, file})
		}
	}

	// Write the stats of each segment, returning any error from writing them
	var w *csv.Writer
	if cmd.CSV != "" {
		f, ferr := os.Create(cmd.CSV)
		if ferr != nil {
			return ferr
		}
		defer func() { err = errors.Join(err, f.Close()) }()
		w = csv.NewWriter(f)
		if err := w.Write(header("label", "path")); err != nil {
			return err
		}
	}

	// Segment the files, until interrupted
	rows, err := cmd.segment(ctx, sources, opts, w)
	if w != nil {
		w.Flush()
		err = errors.Join(err, w.Error())
	}
	if err != nil {
		return err
	}
	summarise(os.Stdout, labels, rows)
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// segment returns the segments of every source, and writes their stats to w
// if it's not nil
func (cmd *ClassifyCmd) segment(ctx server.Cmd, sources []source, opts []segmenter.Opt, w *csv.Writer) ([]row, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var rows []row
	var done int
	var result error

	jobs := cmd.Jobs
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}
	ch := make(chan source)
	for range jobs {
		wg.Go(func() {
			for src := range ch {
				segments, err := decode(ctx.Context(), src.path, cmd.Rate, opts)

				mu.Lock()
				done++
				if err != nil {
					ctx.Logger().Warn("decode", "path", src.path, "error", err)
				}
				for _, s := range segments {
					rows = append(rows, row{src, s})
					if w != nil && result == nil {
						result = w.Write(record(s, src.label, src.path))
					}
				}
				fmt.Fprintf(os.Stderr, "\r%d/%d files", done, len(sources))
				mu.Unlock()
			}
		})
	}
	for _, src := range sources {
		if ctx.Context().Err() != nil {
			break
		}
		ch <- src
	}
	close(ch)
	wg.Wait()
	fmt.Fprintln(os.Stderr)

	if result == nil {
		result = ctx.Context().Err()
	}
	return rows, result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// find returns the audio files in a folder, or up to limit of them spread
// across the folder
func find(path string, limit int) ([]string, error) {
	var files []string
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && slices.Contains(extensions, strings.ToLower(filepath.Ext(path))) {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	slices.Sort(files)
	if limit <= 0 || len(files) <= limit {
		return files, nil
	}
	result := make([]string, limit)
	for i := range result {
		result[i] = files[i*len(files)/limit]
	}
	return result, nil
}

// decode returns the segments of a file, without their samples
func decode(ctx context.Context, path string, rate int, opts []segmenter.Opt) ([]segmenter.Segment, error) {
	seg, err := segmenter.Open(path, rate, opts...)
	if err != nil {
		return nil, err
	}
	defer seg.Close()

	var result []segmenter.Segment
	err = seg.Decode(ctx, func(s *segmenter.Segment) error {
		result = append(result, segmenter.Segment{Start: s.Start, Duration: s.Duration, Class: s.Class, Stats: s.Stats})
		return nil
	})
	return result, err
}

// summarise writes how the segments of each label were classified, weighted
// by duration, and the percentiles of each feature for segments which aren't
// silence
func summarise(w io.Writer, labels []string, rows []row) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	defer tw.Flush()

	fmt.Fprint(tw, "label\tsegments\thours\t")
	for _, c := range classes {
		fmt.Fprintf(tw, "%s\t", className(c))
	}
	fmt.Fprintln(tw)
	for _, label := range labels {
		var count int
		var total time.Duration
		durations := make(map[segmenter.Class]time.Duration)
		for _, r := range rows {
			if r.label == label {
				count++
				total += r.Duration
				durations[r.Class] += r.Duration
			}
		}
		fmt.Fprintf(tw, "%s\t%d\t%.1f\t", label, count, total.Hours())
		for _, c := range classes {
			fmt.Fprintf(tw, "%.1f%%\t", percent(durations[c], total))
		}
		fmt.Fprintln(tw)
	}

	fmt.Fprintln(tw, "\t")
	fmt.Fprint(tw, "feature (p10 / p50 / p90)\t")
	for _, label := range labels {
		fmt.Fprintf(tw, "%s\t", label)
	}
	fmt.Fprintln(tw)
	for _, f := range features {
		fmt.Fprintf(tw, "%s\t", f.name)
		for _, label := range labels {
			var values []float64
			for _, r := range rows {
				if r.label == label && r.Class != segmenter.ClassSilence {
					values = append(values, f.value(r.Stats))
				}
			}
			slices.Sort(values)
			fmt.Fprintf(tw, "%.3f / %.3f / %.3f\t", percentile(values, 10), percentile(values, 50), percentile(values, 90))
		}
		fmt.Fprintln(tw)
	}
}

func percent(d, total time.Duration) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(d) / float64(total)
}

// percentile returns the pth percentile of sorted values
func percentile(values []float64, p int) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[min(len(values)-1, p*len(values)/100)]
}
