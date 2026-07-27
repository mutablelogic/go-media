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
	GetAudioProfile    GetAudioProfileCmd    `cmd:"" name:"audio-profile" help:"Get the details of an audio profile." group:"PROFILE"`
	DeleteAudioProfile DeleteAudioProfileCmd `cmd:"" name:"audio-profile-delete" help:"Delete an audio profile." group:"PROFILE"`
	UpdateAudioProfile UpdateAudioProfileCmd `cmd:"" name:"audio-profile-update" help:"Update an audio profile." group:"PROFILE"`
}

type CreateAudioProfileCmd struct {
	schema.AudioProfileMeta
}

type GetAudioProfileCmd struct {
	UUID string `arg:"" name:"uuid" help:"UUID of the audio profile."`
}

type DeleteAudioProfileCmd struct {
	UUID string `arg:"" name:"uuid" help:"UUID of the audio profile."`
}

type UpdateAudioProfileCmd struct {
	UUID string `arg:"" name:"uuid" help:"UUID of the audio profile."`
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

func (cmd *GetAudioProfileCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "GetAudioProfile", func(ctx context.Context, client *httpclient.Client) error {
		// Get the audio profile
		profile, err := client.GetAudioProfile(ctx, cmd.UUID)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(profile))

		// Return success
		return nil
	})
}

func (cmd *DeleteAudioProfileCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "DeleteAudioProfile", func(ctx context.Context, client *httpclient.Client) error {
		// Delete the audio profile
		profile, err := client.DeleteAudioProfile(ctx, cmd.UUID)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(profile))

		// Return success
		return nil
	})
}

func (cmd *UpdateAudioProfileCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "UpdateAudioProfile", func(ctx context.Context, client *httpclient.Client) error {
		// Update the audio profile
		profile, err := client.UpdateAudioProfile(ctx, cmd.UUID, cmd.AudioProfileMeta)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(profile))

		// Return success
		return nil
	})
}
