package main

import (
	"fmt"
	"os"

	// Packages
	chromaprintschema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type FingerprintCommands struct {
	AudioFingerprint AudioFingerprintCommand `cmd:"" help:"Audio fingerprint (and lookup)" group:"PROCESS"`
}

type AudioFingerprintCommand struct {
	chromaprintschema.AudioFingerprintRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *AudioFingerprintCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method
	response, err := globals.manager.AudioFingerprint(globals.ctx, &cmd.AudioFingerprintRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}
