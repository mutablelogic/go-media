package main

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"

	// Packages
	thumbnailer "github.com/mutablelogic/go-media/thumbnailer"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// ThumbnailOpts are the options for finding thumbnails
type ThumbnailOpts struct {
	Width   int    `name:"width" default:"640" help:"Width of thumbnails, or zero to keep the aspect ratio with --height, or the video's own size if both are zero."`
	Height  int    `name:"height" help:"Height of thumbnails, or zero to keep the aspect ratio."`
	Quality int    `name:"quality" default:"85" help:"JPEG quality (1-100)."`
	Mode    string `name:"mode" enum:"auto,keyframes,seek" default:"auto" help:"How frames are sampled: auto, keyframes, or seek to evenly spaced times."`
	Samples int    `name:"samples" default:"300" help:"Maximum number of frames sampled by seeking."`
	Stream  *int   `name:"stream" help:"Index of the video stream (default the first)."`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var modes = map[string]thumbnailer.Mode{
	"auto":      thumbnailer.ModeAuto,
	"keyframes": thumbnailer.ModeKeyframes,
	"seek":      thumbnailer.ModeSeek,
}

// rowFormat is the format of each row printed for a thumbnail or scene
const rowFormat = "%-6v  %-12v  %-12v  %-6v  %5v  %v\n"

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// open opens a video with the options
func (o ThumbnailOpts) open(path string) (*thumbnailer.Thumbnailer, error) {
	if o.Quality < 1 || o.Quality > 100 {
		return nil, fmt.Errorf("invalid quality %d", o.Quality)
	}
	mode, exists := modes[o.Mode]
	if !exists {
		return nil, fmt.Errorf("invalid mode %q", o.Mode)
	}
	opts := []thumbnailer.Opt{
		thumbnailer.WithSize(o.Width, o.Height),
		thumbnailer.WithMode(mode),
		thumbnailer.WithMaxSamples(o.Samples),
	}
	if o.Stream != nil {
		opts = append(opts, thumbnailer.WithStream(*o.Stream))
	}
	return thumbnailer.Open(path, opts...)
}

// write writes an image as a JPEG file
func (o ThumbnailOpts) write(path string, img image.Image) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: o.Quality})
}

// writeAll writes images as JPEG files in a folder, named by their number
// and time, and returns their paths, or nothing if the folder is empty
func (o ThumbnailOpts) writeAll(folder string, times []time.Duration, images []image.Image) ([]string, error) {
	paths := make([]string, len(images))
	if folder == "" {
		return paths, nil
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return nil, err
	}
	for i, img := range images {
		paths[i] = filepath.Join(folder, fmt.Sprintf("%03d-%.3fs.jpg", i+1, times[i].Seconds()))
		if err := o.write(paths[i], img); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// round returns a duration rounded to the nearest 10ms
func round(d time.Duration) time.Duration {
	return d.Round(10 * time.Millisecond)
}
