package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	file "github.com/mutablelogic/go-media/pkg/file"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataCommands struct {
	Meta       ListMetadata      `cmd:"" group:"METADATA" help:"Examine metadata"`
	Artwork    ExtractArtwork    `cmd:"" group:"METADATA" help:"Extract artwork"`
	Streams    ListStreams       `cmd:"" group:"METADATA" help:"List streams"`
	Thumbnails ExtractThumbnails `cmd:"" group:"METADATA" help:"Extract video thumbnails"`
}

type ListMetadata struct {
	Path      string `arg:"" type:"path" help:"File or directory"`
	Recursive bool   `short:"r" help:"Recursively examine files"`
}

type ExtractArtwork struct {
	Path      string `arg:"" type:"path" help:"File or directory"`
	Recursive bool   `short:"r" help:"Recursively examine files"`
	Out       string `required:"" help:"Output filename for artwork, relative to the source path. Use {count} {hash} {path} {name} or {ext} for placeholders" default:"{hash}{ext}"`
}

type ListStreams struct {
	Path      string `arg:"" type:"path" help:"File or directory"`
	Recursive bool   `short:"r" help:"Recursively examine files"`
}

type ExtractThumbnails struct {
	Path  string        `arg:"" type:"path" help:"File"`
	Out   string        `required:"" help:"Output filename for thumbnail, relative to the source path. Use {timestamp} {frame} {path} {name} or {ext} for placeholders" default:"{frame}{ext}"`
	N     uint64        `short:"n" help:"Maxumum number of frames to extract"`
	Delta time.Duration `short:"d" help:"Time between frames" default:"1s"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListStreams) Run(app server.Cmd) error {
	// Create the media manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(false, nil))
	if err != nil {
		return err
	}

	// Create a new file walker
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info os.FileInfo) error {
		if info.IsDir() {
			if !cmd.Recursive && relpath != "." {
				return file.SkipDir
			}
			return nil
		}

		// Open file
		f, err := manager.Open(filepath.Join(root, relpath), nil)
		if err != nil {
			return fmt.Errorf("%s: %w", info.Name(), err)
		}
		defer f.Close()

		// Enumerate streams
		streams := f.(*ffmpeg.Reader).Streams(media.ANY)
		result := make([]media.Metadata, 0, len(streams))
		result = append(result, ffmpeg.NewMetadata("path", filepath.Join(root, relpath)))
		for _, meta := range streams {
			result = append(result, ffmpeg.NewMetadata(fmt.Sprint(meta.Index()), meta))
		}

		return write(os.Stdout, result, nil)
	})

	// Perform the walk, return any errors
	return walker.Walk(app.Context(), cmd.Path)
}

func (cmd *ListMetadata) Run(app server.Cmd) error {
	// Create the media manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(false, nil))
	if err != nil {
		return err
	}

	// Create a new file walker
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info os.FileInfo) error {
		if info.IsDir() {
			if !cmd.Recursive && relpath != "." {
				return file.SkipDir
			}
			return nil
		}

		// Open file
		f, err := manager.Open(filepath.Join(root, relpath), nil)
		if err != nil {
			return fmt.Errorf("%s: %w", info.Name(), err)
		}
		defer f.Close()

		// Print metadata
		result := make([]media.Metadata, 0, 20)
		result = append(result, ffmpeg.NewMetadata("path", filepath.Join(root, relpath)))
		result = append(result, ffmpeg.NewMetadata("type", f.Type()))

		if duration := f.(*ffmpeg.Reader).Duration(); duration > 0 {
			result = append(result, ffmpeg.NewMetadata("duration", duration.String()))
		}

		for _, meta := range f.(*ffmpeg.Reader).Metadata() {
			result = append(result, meta)
		}

		return write(os.Stdout, result, nil)
	})

	// Perform the walk, return any errors
	return walker.Walk(app.Context(), cmd.Path)
}

func (cmd *ExtractArtwork) Run(app server.Cmd) error {
	// Create the media manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(false, nil))
	if err != nil {
		return err
	}

	// Create a new file walker
	walker := file.NewWalker(func(ctx context.Context, root, relpath string, info os.FileInfo) error {
		if info.IsDir() {
			if !cmd.Recursive && relpath != "." {
				return file.SkipDir
			}
			return nil
		}

		// Open file
		f, err := manager.Open(filepath.Join(root, relpath), nil)
		if err != nil {
			return fmt.Errorf("%s: %w", info.Name(), err)
		}
		defer f.Close()

		// Extract artwork
		count := 1
		result := make([]media.Metadata, 0, 20)
		for _, meta := range f.(*ffmpeg.Reader).Metadata("artwork") {
			data := meta.Bytes()
			mimetype, ext, err := file.MimeType(data)
			if err != nil {
				return err
			}

			// Determine the output filename
			out := template(cmd.Out, "hash", types.Hash(data), "path", filepath.Dir(relpath), "name", info.Name(), "mimetype", mimetype, "ext", ext, "count", count)

			// If the filename is relative, make it absolute
			if !filepath.IsAbs(out) {
				out = filepath.Join(root, out)
			}

			// If file exists, skip it
			if stat, err := os.Stat(out); err == nil && stat.Mode().IsRegular() && stat.Size() == int64(len(data)) {
				continue
			}

			// Make the directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Writing %s\n", out)
			if err := os.WriteFile(out, data, 0644); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			count++
		}

		return write(os.Stdout, result, nil)
	})

	// Perform the walk, return any errors
	return walker.Walk(app.Context(), cmd.Path)
}

func (cmd *ExtractThumbnails) Run(app server.Cmd) error {
	// Create the media manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(false, nil))
	if err != nil {
		return err
	}

	// Open file
	input, err := manager.Open(cmd.Path, nil)
	if err != nil {
		return err
	}
	defer input.Close()

	mapfunc := func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == input.(*ffmpeg.Reader).BestStream(media.VIDEO) {
			// Convert frame to yuv420p if needed, but use the same size and frame rate
			return ffmpeg.NewVideoPar("yuv420p", par.WidthHeight(), par.FrameRate())
		}
		// Ignore other streams
		return nil, nil
	}

	// Decode the streams and receive the video frame
	nframe := 0
	var oldts time.Duration
	return input.(*ffmpeg.Reader).Decode(app.Context(), mapfunc, func(stream int, frame *ffmpeg.Frame) error {
		nframe++
		ts := time.Duration(frame.Ts()*1000) * time.Millisecond
		if ts < 0 {
			// Packets being flushed?
			return nil
		}

		// Skip if ts is too soon after the last frame
		if cmd.Delta > 0 && oldts > 0 && ts-oldts < cmd.Delta {
			return nil
		} else {
			oldts = ts
		}

		// Determine the output filename
		out := template(cmd.Out, "frame", nframe, "ext", ".jpg", "timestamp", ts)

		// Make the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}

		// Write the frame to a file
		fmt.Println("Writing", out)
		w, err := os.Create(out)
		if err != nil {
			return err
		}
		defer w.Close()

		// Convert to an image and encode a JPEG
		if image, err := frame.Image(); err != nil {
			return err
		} else if err := jpeg.Encode(w, image, nil); err != nil {
			return err
		}

		// Check maximum number of frames
		if cmd.N > 0 && uint64(nframe) >= cmd.N {
			return io.EOF
		}

		// Seek to the next frame - actually the keyframe just
		// before the next frame, so we can skip a few frames
		if cmd.Delta > 0 {
			newts := ts + cmd.Delta
			if newts < input.(*ffmpeg.Reader).Duration() {
				if err := input.(*ffmpeg.Reader).Seek(stream, newts.Truncate(time.Second).Seconds()); err != nil {
					return err
				}
			}
		}

		// Return for next frame
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func template(tmpl string, args ...any) string {
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			tmpl = strings.ReplaceAll(tmpl, fmt.Sprintf("{%s}", args[i]), fmt.Sprint(args[i+1]))
		}
	}
	return tmpl
}
