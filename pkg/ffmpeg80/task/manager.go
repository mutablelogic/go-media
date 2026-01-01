package task

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	opts
	chromaprint *chromaprint.Client
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(opt ...Opt) (*Manager, error) {
	m := new(Manager)
	if err := applyOpts(&m.opts, opt...); err != nil {
		return nil, err
	}

	// If there is a trace function, set it
	if m.tracefn != nil {
		ffmpeg.SetLogging(m.verbose, ffmpeg.LogFn(m.tracefn))
	}

	// If a chromaprint key is set, initialize chromaprint
	if m.chromaprintKey != "" {
		clientopts := []client.ClientOpt{}
		if client, err := chromaprint.NewClient(m.chromaprintKey, clientopts...); err != nil {
			return nil, err
		} else {
			m.chromaprint = client
		}
	}

	// Success
	return m, nil
}
