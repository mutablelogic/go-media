package main

import (
	"fmt"
	"os"

	// Packages
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type FingerprintCommands struct {
	MatchMusic MatchMusic `cmd:"" group:"MATCH" help:"Match Music Track"`
}

type MatchMusic struct {
	Path   string   `arg:"" type:"path" help:"File"`
	APIKey string   `env:"CHROMAPRINT_KEY" help:"API key for the music matching service (https://acoustid.org/login)"`
	Type   []string `cmd:"" help:"Type of match to perform" enum:"any,recording,release,releasegroup,track" default:"any"`
	Score  float64  `cmd:"" help:"Minimum match scoreto perform" default:"0.9"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *MatchMusic) Run(app server.Cmd) error {
	ffmpeg.SetLogging(false, nil)

	// Create a client
	client, err := chromaprint.NewClient(cmd.APIKey)
	if err != nil {
		return err
	}

	// Open the file
	r, err := os.Open(cmd.Path)
	if err != nil {
		return err
	}
	defer r.Close()

	var meta chromaprint.Meta
	for _, t := range cmd.Type {
		switch t {
		case "any":
			meta |= chromaprint.META_ALL
		case "recording":
			meta |= chromaprint.META_RECORDING
		case "release":
			meta |= chromaprint.META_RELEASE
		case "releasegroup":
			meta |= chromaprint.META_RELEASEGROUP
		case "track":
			meta |= chromaprint.META_TRACK
		default:
			return fmt.Errorf("unknown type %q", t)
		}
	}

	// Create the matches
	matches, err := client.Match(app.Context(), r, meta)
	if err != nil {
		return err
	}

	// Filter by score
	result := make([]*chromaprint.ResponseMatch, 0, len(matches))
	for _, m := range matches {
		if m.Score >= cmd.Score {
			result = append(result, m)
		}
	}

	fmt.Println(result)
	return nil
}

/*

	META_RECORDING Meta = (1 << iota)
	META_RECORDINGID
	META_RELEASE
	META_RELEASEID
	META_RELEASEGROUP
	META_RELEASEGROUPID
	META_TRACK
	META_COMPRESS
	META_USERMETA
	META_SOURCE
*/
