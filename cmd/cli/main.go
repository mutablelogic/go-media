package main

import (
	"os"
	"path/filepath"

	// Packages
	kong "github.com/alecthomas/kong"
)

type Globals struct {
	Debug bool `name:"debug" help:"Enable debug mode"`
}

type CLI struct {
	Globals
	Version        VersionCmd        `cmd:"version" help:"Print version information"`
	Demuxers       DemuxersCmd       `cmd:"demuxers" help:"List media demultiplex (input) formats"`
	Muxers         MuxersCmd         `cmd:"muxers" help:"List media multiplex (output) formats"`
	SampleFormats  SampleFormatsCmd  `cmd:"samplefmts" help:"List audio sample formats"`
	ChannelLayouts ChannelLayoutsCmd `cmd:"channellayouts" help:"List audio channel layouts"`
	PixelFormats   PixelFormatsCmd   `cmd:"pixelfmts" help:"List video pixel formats"`
	Metadata       MetadataCmd       `cmd:"metadata" help:"Display media metadata information"`
	Probe          ProbeCmd          `cmd:"probe" help:"Probe media file or device"`
	Decode         DecodeCmd         `cmd:"decode" help:"Decode media"`
}

func main() {
	name, err := os.Executable()
	if err != nil {
		panic(err)
	}

	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name(filepath.Base(name)),
		kong.Description("commands for media processing"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
	)
	if err := ctx.Run(&cli.Globals); err != nil {
		ctx.FatalIfErrorf(err)
	}
}
