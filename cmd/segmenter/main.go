// segmenter cuts audio into segments at silence, at a fixed duration, or
// near given times, and classifies each segment as silence, speech, music or
// other sound. It runs locally, without a server.
//
//	segmenter segment --segments 3m10s,7m45s --format mp3 podcast.mp3 parts
//	segmenter stream device://avfoundation/:0
//	segmenter classify --limit 50 --csv stats.csv musan/speech musan/music musan/noise
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
	Segment  SegmentCmd  `cmd:"" help:"Cut audio into segments, print them, and optionally write each one to a file."`
	Stream   StreamCmd   `cmd:"" help:"Cut a stream or device, such as a microphone, into segments as it plays, showing its level on a terminal."`
	Classify ClassifyCmd `cmd:"" help:"Classify folders of labelled audio, and summarise how each label was classified."`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	if err := cmd.Main(CLI{}, "segmenter cuts audio into segments, and classifies them as silence, speech, music or other sound", version.Version()); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
