package main

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	kong "github.com/alecthomas/kong"
	tablewriter "github.com/djthorpe/go-tablewriter"
	media "github.com/mutablelogic/go-media"
)

type Globals struct {
	manager media.Manager
	writer  *tablewriter.Writer
	ctx     context.Context

	Debug bool `name:"debug" help:"Enable debug mode"`
	Force bool `name:"force" help:"Force resampling and resizing on decode, even if the input and output parameters are the same"`
}

type CLI struct {
	Globals
	Version        VersionCmd        `cmd:"version" help:"Print version information"`
	Demuxers       DemuxersCmd       `cmd:"demuxers" help:"List media demultiplex (input) formats"`
	Muxers         MuxersCmd         `cmd:"muxers" help:"List media multiplex (output) formats"`
	Devices        DevicesCmd        `cmd:"devices" help:"List inout and output devices"`
	Codecs         CodecsCmd         `cmd:"codecs" help:"List audio and video codecs"`
	SampleFormats  SampleFormatsCmd  `cmd:"samplefmts" help:"List audio sample formats"`
	ChannelLayouts ChannelLayoutsCmd `cmd:"channellayouts" help:"List audio channel layouts"`
	PixelFormats   PixelFormatsCmd   `cmd:"pixelfmts" help:"List video pixel formats"`
	Metadata       MetadataCmd       `cmd:"metadata" help:"Display media metadata information"`
	Artwork        ArtworkCmd        `cmd:"artwork" help:"Save artwork from media file"`
	Probe          ProbeCmd          `cmd:"probe" help:"Probe media file or device"`
	Decode         DecodeCmd         `cmd:"decode" help:"Decode media"`
	Thumbnails     ThumbnailsCmd     `cmd:"thumbnails" help:"Generate thumbnails from media file"`
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

	// Create a manager object
	// Only print out FATAL messages
	manager, err := media.NewManager(media.OptLog(cli.Debug, nil))
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	cli.Globals.manager = manager

	// Create a tablewriter object with text output
	writer := tablewriter.New(os.Stdout, tablewriter.OptOutputText())
	cli.Globals.writer = writer

	// Create a context
	cli.Globals.ctx = ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Run the command
	if err := ctx.Run(&cli.Globals); err != nil {
		ctx.FatalIfErrorf(err)
	}
}
