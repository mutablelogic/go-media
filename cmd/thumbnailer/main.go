// thumbnailer finds the scenes of a video, and representative thumbnails of
// it, which it writes as JPEG files. It runs locally, without a server.
//
//	thumbnailer thumbs --count 6 video.mp4 thumbs
//	thumbnailer poster video.mp4 poster.jpg
//	thumbnailer scenes video.mp4 scenes
package main

import (
	"fmt"
	"os"

	// Packages
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	Thumbs ThumbsCmd `cmd:"" help:"Choose thumbnails from different scenes of a video, print them, and optionally write each one as a JPEG file."`
	Poster PosterCmd `cmd:"" help:"Choose the thumbnail which best represents a video, print it, and optionally write it as a JPEG file."`
	Scenes ScenesCmd `cmd:"" help:"Find the scenes of a video, print them, and optionally write the best frame of each as a JPEG file."`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	if err := cmd.Main(CLI{}, "thumbnailer finds the scenes of a video, and representative thumbnails of it", version.Version()); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
