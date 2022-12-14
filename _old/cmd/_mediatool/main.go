package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	// Packages
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Command struct {
	*flag.FlagSet
	Keyword     string
	Syntax      string
	Description string
	Fn          func(context.Context, *Command, []string) error
	Errs        chan error
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	commands = []Command{}
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

func main() {
	flags := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Syntax:")
		fmt.Fprintln(flags.Output(), "  ", flags.Name(), "<flags> command args...")
		fmt.Fprintln(flags.Output(), "")
		fmt.Fprintln(flags.Output(), "Commands:")
		PrintCommands(flags.Output())
		fmt.Fprintln(flags.Output(), "")
		fmt.Fprintln(flags.Output(), "Flags:")
		flags.PrintDefaults()
	}

	// Add commands
	commands = append(commands, RegisterCommands(flags)...)

	// Parse command line flags
	if err := flags.Parse(os.Args[1:]); err == flag.ErrHelp {
		os.Exit(0)
	}
	if flags.NArg() < 1 {
		flags.Usage()
		os.Exit(0)
	}
	if command := GetCommand(flags.Arg(0)); command == nil {
		fmt.Fprintln(os.Stderr, "Unknown command:", flags.Arg(0))
		os.Exit(-1)
	} else if err := RunCommand(flags, command); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func PrintCommands(w io.Writer) {
	for _, command := range commands {
		fmt.Fprintf(w, "  %-10s %s\n", command.Keyword, command.Syntax)
		fmt.Fprintf(w, "    \t%s\n", command.Description)
	}
}

func RegisterCommands(flags *flag.FlagSet) []Command {
	result := []Command{}
	result = append(result, GetHelp, GetVersion, GetMetadata, GetArtwork)
	return result
}

func GetCommand(keyword string) *Command {
	for _, command := range commands {
		if command.Keyword == keyword {
			return &command
		}
	}
	return nil
}

func RunCommand(f *flag.FlagSet, cmd *Command) error {
	// Set up command
	cmd.FlagSet = f
	cmd.Errs = make(chan error)

	// Create context with cancel, cancel called on CTRL+C
	ctx := HandleSignal()

	// Write out errors
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
	FOR_LOOP:
		for {
			select {
			case <-ctx.Done():
				break FOR_LOOP
			case err := <-cmd.Errs:
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}()

	// Run command
	err := cmd.Fn(ctx, cmd, f.Args()[1:])

	// Return any errors
	return err
}

func HandleSignal() context.Context {
	// Handle signals - call cancel when interrupt received
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
