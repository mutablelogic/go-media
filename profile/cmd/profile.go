package cmd

import (
	"context"
	"fmt"

	// Packages
	httpclient "github.com/mutablelogic/go-media/profile/httpclient"
	schema "github.com/mutablelogic/go-media/profile/schema"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientProfileCommands struct {
	CreateAudioProfile CreateAudioProfileCmd `cmd:"" name:"audio-profile-create" help:"Create a new audio profile." group:"PROFILE"`
}

type CreateAudioProfileCmd struct {
	schema.AudioProfileMeta
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *CreateAudioProfileCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "CreateAudioProfile", func(ctx context.Context, client *httpclient.Client) error {
		// Create the audio profile
		profile, err := client.CreateAudioProfile(ctx, cmd.AudioProfileMeta)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(profile))

		// Return success
		return nil
	})
}
