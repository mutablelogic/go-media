package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Flags struct {
	*flag.FlagSet

	in, out                *string
	audio, video, subtitle *bool
	version                *bool
	debug                  *bool
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFlags(args []string) (*Flags, error) {
	flags := &Flags{
		FlagSet: flag.NewFlagSet(filepath.Base(args[0]), flag.ContinueOnError),
	}
	flags.register()
	flags.Usage = func() {
		fmt.Fprintf(flags.Writer(), "\n%s: Media transcoding frontend\n\n", flags.Name())
		fmt.Fprintln(flags.Writer(), "Usage:")
		fmt.Fprintf(flags.Writer(), "  %s -version\n  \tPrint versions and exit\n", flags.Name())
		fmt.Fprintf(flags.Writer(), "  %s -in <format>\n  \tList input file and device formats (use '*' for all)\n", flags.Name())
		fmt.Fprintf(flags.Writer(), "  %s -out <format>\n  \tList output file and device formats (use '*' for all)\n", flags.Name())
		fmt.Fprintf(flags.Writer(), "  %s [options] <input>\n  \tRead input file\n", flags.Name())
		fmt.Fprintln(flags.Writer(), "\nOptions:")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args[1:]); errors.Is(err, flag.ErrHelp) {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	// Return success
	return flags, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (flags *Flags) Writer() io.Writer {
	return flags.FlagSet.Output()
}

func (flags *Flags) Version() bool {
	return *flags.version
}

func (flags *Flags) Debug() bool {
	return *flags.debug
}

func (flags *Flags) In() string {
	return *flags.in
}

func (flags *Flags) Out() string {
	return *flags.out
}

func (flags *Flags) MediaFlags() MediaFlag {
	media_type := MEDIA_FLAG_NONE
	if *flags.audio {
		media_type |= MEDIA_FLAG_AUDIO
	}
	if *flags.video {
		media_type |= MEDIA_FLAG_VIDEO
	}
	if *flags.subtitle {
		media_type |= MEDIA_FLAG_SUBTITLE
	}
	return media_type
}

func (flags *Flags) PrintShortUsage() {
	fmt.Fprintf(flags.Writer(), "%s: missing argument, try -help for more information\n", flags.Name())
}

func (flags *Flags) PrintFormats(title string, formats []MediaFormat) {
	fmt.Fprintln(flags.Writer(), title)
	for _, format := range formats {
		flag_str := strings.ToLower(strings.Replace(fmt.Sprint(format.Flags()), "MEDIA_FLAG_", "", -1))
		fmt.Fprintf(flags.Writer(), "  %s (%s)\n", format.Description(), flag_str)
		if name := format.Name(); len(name) > 0 {
			fmt.Fprintf(flags.Writer(), "  \tName: %s\n", strings.Join(name, ", "))
		}
		if ext := format.Ext(); len(ext) > 0 {
			fmt.Fprintf(flags.Writer(), "  \tExt: %s\n", strings.Join(ext, ", "))
		}
		if mimetype := format.MimeType(); len(mimetype) > 0 {
			fmt.Fprintf(flags.Writer(), "  \tMimeType: %s\n", strings.Join(mimetype, ", "))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (flags *Flags) register() {
	flags.version = flags.Bool("version", false, "Print version information")
	flags.debug = flags.Bool("debug", false, "Enable debug output")
	flags.in = flags.String("in", "", "Input format. If not specified, input format is auto-detected")
	flags.out = flags.String("out", "", "Output filename or name")
	flags.audio = flags.Bool("audio", false, "Extract audio")
	flags.video = flags.Bool("video", false, "Extract video")
	flags.subtitle = flags.Bool("subtitle", false, "Extract subtitles")
}
