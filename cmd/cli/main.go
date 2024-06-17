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
	Version  VersionCmd  `cmd:"version" help:"Print version information"`
	Metadata MetadataCmd `cmd:"metadata" help:"Display media metadata information"`
	Decode   DecodeCmd   `cmd:"decode" help:"Decode media"`
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
