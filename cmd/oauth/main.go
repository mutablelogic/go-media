package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	// Packages
	googleclient "github.com/mutablelogic/go-media/pkg/googleclient"
	googlephotos "github.com/mutablelogic/go-media/pkg/googlephotos"
)

////////////////////////////////////////////////////////////////////////////////
// FLAGS

var (
	flagCommandLineAuth = flag.String("client_secret", "", "Client Secret")
	flagCommandLineCode = flag.String("code", "", "authentication code")
	flagScope           = flag.String("scope", "", "Comma-separated list of scopes")
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

func main() {
	flag.Parse()

	// Context with cancel
	ctx := HandleSignal()

	// Create a client
	client, err := googleclient.NewClientWithClientSecret(googleclient.Config{
		Name:   "googlephotos",
		Scopes: strings.Fields(*flagScope),
	}, *flagCommandLineAuth)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	}

	// Read cached token
	token, err := client.ReadToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	}

	if token == nil {
		// No token is available
		if *flagCommandLineCode != "" {
			token, err = client.CommandLineToken(ctx, *flagCommandLineCode)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(-1)
			}
			if err := client.WriteToken(token); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(-1)
			}
		} else {
			auth, err := client.CommandLineAuth()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(-1)
			}
			fmt.Println("Navigate to", auth.VerificationURL, "and use -code flag to enter the code to create a token")
		}
	}

	// Set the oauth token
	if err := client.Use(ctx, token, googlephotos.Endpoint); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	}

	// Retrieve mediaitems
	var result googlephotos.Array
	if err := client.Get("/v1/mediaItems", &result); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	}
	fmt.Println(result.MediaItems)
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
