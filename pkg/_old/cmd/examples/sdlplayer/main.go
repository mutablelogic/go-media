/* This example demonstrates how to play audio and video files using SDL2. */
package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func main() {
	// Bail out when we receive a signal
	ctx := ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Create a player object
	player := NewPlayer()
	defer player.Close()

	// Open the file
	var result error
	if len(os.Args) == 2 {
		result = player.OpenUrl(os.Args[1])
	} else {
		result = errors.New("usage: sdlplayer <filename>")
	}
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
		os.Exit(-1)
	}

	// Play
	if err := player.Play(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
