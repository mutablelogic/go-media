package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	// Packages
	"github.com/mutablelogic/go-media/pkg/googleclient"
	"github.com/mutablelogic/go-media/pkg/googlephotos"
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
	commands     = []Command{}
	clientsecret string
	scope        string
	authcode     string
	debug        bool
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

	// Define global flags
	DefineFlags(flags)

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
	result = append(result, GetHelp, GetVersion, GetAlbums, GetSharedAlbums, GetMedia)
	return result
}

func DefineFlags(flags *flag.FlagSet) {
	// Define global flags
	flags.StringVar(&clientsecret, "clientsecret", "client_secret.json", "Client secret")
	flags.StringVar(&scope, "scope", googlephotos.Scope, "Comma-separated list of scopes")
	flags.StringVar(&authcode, "auth", "", "Authentication code")
	flags.BoolVar(&debug, "debug", false, "Debug mode")
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

////////////////////////////////////////////////////////////////////////////////
// AUTHENTICATION FLOW

func (c *Command) Authenticate(ctx context.Context) (*googlephotos.Client, error) {
	// Create a client
	client, err := googleclient.NewClientWithClientSecret(googleclient.Config{
		Name:   c.Name(),
		Scopes: strings.Fields(scope),
	}, clientsecret)
	if err != nil {
		return nil, err
	}

	// Read cached token
	token, err := client.ReadToken()
	if err != nil {
		return nil, err
	}

	// No token is available - auth flow
	if token == nil {
		if authcode != "" {
			token, err = client.CommandLineToken(ctx, authcode)
			if err != nil {
				return nil, err
			}
			if err := client.WriteToken(token); err != nil {
				return nil, err
			}
		} else {
			auth, err := client.CommandLineAuth()
			if err != nil {
				return nil, err
			}
			fmt.Println("Navigate to", auth.VerificationURL, "and use -auth flag to enter the code to create a token")
		}
	}

	// Set the oauth token
	if err := client.Use(ctx, token, googlephotos.Endpoint); err != nil {
		return nil, err
	}

	// Return success
	return googlephotos.NewClient(client), nil
}
