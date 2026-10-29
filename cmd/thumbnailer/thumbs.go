package main

import (
	"fmt"
	"image"
	"path/filepath"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ThumbsCmd struct {
	ThumbnailOpts
	Path   string `arg:"" name:"path" help:"Path or URL of the video."`
	Folder string `arg:"" optional:"" name:"folder" type:"path" help:"Folder to write each thumbnail to, as a JPEG file."`
	Count  int    `name:"count" short:"n" default:"6" help:"Number of thumbnails."`
}

type PosterCmd struct {
	ThumbnailOpts
	Path string `arg:"" name:"path" help:"Path or URL of the video."`
	File string `arg:"" optional:"" name:"file" type:"path" help:"JPEG file to write the poster to."`
}

type ScenesCmd struct {
	ThumbnailOpts
	Path   string `arg:"" name:"path" help:"Path or URL of the video."`
	Folder string `arg:"" optional:"" name:"folder" type:"path" help:"Folder to write the best frame of each scene to, as a JPEG file."`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ThumbsCmd) Run(ctx server.Cmd) error {
	if cmd.Count < 1 {
		return fmt.Errorf("invalid count %d", cmd.Count)
	}
	th, err := cmd.open(cmd.Path)
	if err != nil {
		return err
	}
	defer th.Close()

	thumbs, err := th.Thumbnails(ctx.Context(), cmd.Count)
	if err != nil {
		return err
	}
	times, images := make([]time.Duration, len(thumbs)), make([]image.Image, len(thumbs))
	for i, thumb := range thumbs {
		times[i], images[i] = thumb.Time, thumb.Image
	}
	paths, err := cmd.writeAll(cmd.Folder, times, images)
	if err != nil {
		return err
	}

	fmt.Printf(rowFormat, "THUMB", "TIME", "SIZE", "SCENE", "SCORE", "FILE")
	for i, thumb := range thumbs {
		fmt.Printf(rowFormat, i+1, round(thumb.Time), size(thumb.Image), thumb.Scene+1, fmt.Sprintf("%.2f", thumb.Stats.Score), paths[i])
	}
	return nil
}

func (cmd *PosterCmd) Run(ctx server.Cmd) error {
	th, err := cmd.open(cmd.Path)
	if err != nil {
		return err
	}
	defer th.Close()

	poster, err := th.Poster(ctx.Context())
	if err != nil {
		return err
	}
	if cmd.File != "" {
		if filepath.Ext(cmd.File) == "" {
			cmd.File += ".jpg"
		}
		if err := cmd.write(cmd.File, poster.Image); err != nil {
			return err
		}
	}

	fmt.Printf(rowFormat, "THUMB", "TIME", "SIZE", "SCENE", "SCORE", "FILE")
	fmt.Printf(rowFormat, 1, round(poster.Time), size(poster.Image), poster.Scene+1, fmt.Sprintf("%.2f", poster.Stats.Score), cmd.File)
	return nil
}

func (cmd *ScenesCmd) Run(ctx server.Cmd) error {
	th, err := cmd.open(cmd.Path)
	if err != nil {
		return err
	}
	defer th.Close()

	analysis, err := th.Analyse(ctx.Context())
	if err != nil {
		return err
	}
	ctx.Logger().Debug("analysed", "mode", analysis.Mode, "duration", analysis.Duration, "samples", len(analysis.Samples), "scenes", len(analysis.Scenes))

	// The best frame of each scene, which isn't blank
	var times []time.Duration
	var images []image.Image
	if cmd.Folder != "" {
		for _, scene := range analysis.Scenes {
			if scene.Best < 0 {
				continue
			}
			at := analysis.Samples[scene.Best].Time
			img, err := th.Frame(ctx.Context(), at)
			if err != nil {
				return err
			}
			times, images = append(times, at), append(images, img)
		}
	}
	paths, err := cmd.writeAll(cmd.Folder, times, images)
	if err != nil {
		return err
	}

	fmt.Printf(rowFormat, "SCENE", "START", "DURATION", "HOW", "SCORE", "FILE")
	var n int
	for i, scene := range analysis.Scenes {
		var score, path string
		if scene.Best >= 0 {
			score = fmt.Sprintf("%.2f", analysis.Samples[scene.Best].Stats.Score)
			if n < len(paths) {
				path = paths[n]
				n++
			}
		}
		fmt.Printf(rowFormat, i+1, round(scene.Start), round(scene.Duration), scene.Transition, score, path)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// size returns the size of an image as text
func size(img image.Image) string {
	return fmt.Sprintf("%dx%d", img.Bounds().Dx(), img.Bounds().Dy())
}
