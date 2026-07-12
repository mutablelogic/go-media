//go:build chromaprint

package cmd

import (
	"fmt"
	"os"

	// Packages
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MetadataChromaprintCLICommands struct {
	AudioFingerprint AudioFingerprintCmd `cmd:"" name:"audio-fingerprint" help:"Generate audio fingerprint." group:"METADATA"`
}

type AudioFingerprintCmd struct {
	BaseCmd
	File        string   `arg:"" name:"file" type:"existingfile" help:"File to fingerprint."`
	InputFormat string   `name:"input-format" help:"Input format name (e.g. s16le)"`
	InputOpts   []string `name:"input-opt" help:"Input format option key=value (repeatable)"`
	Duration    float64  `flag:"" name:"duration" help:"Duration in seconds (0 = auto-detect)." default:"0"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *AudioFingerprintCmd) Run(ctx server.Cmd) error {
	json, _ := c.IsJSONOutput(ctx)
	return c.WithManager(ctx, func(manager *manager.Media) error {
		resp, err := manager.AudioFingerprint(ctx.Context(), schema.AudioFingerprintRequest{
			Input:       c.File,
			InputFormat: c.InputFormat,
			InputOpts:   c.InputOpts,
			Duration:    c.Duration,
		})
		if err != nil {
			return err
		}

		if json {
			fmt.Println(resp)
			return nil
		}

		fmt.Fprintf(os.Stdout, "Fingerprint: %s\n", resp.Fingerprint)
		fmt.Fprintf(os.Stdout, "Duration: %.3fs\n", resp.Duration)
		return nil
	})
}
