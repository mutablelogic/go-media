package main

import (
	"context"
	"fmt"
	"os"

	// Packages
	tablewriter "github.com/olekukonko/tablewriter"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GetAlbums = Command{
		Keyword:     "albums",
		Syntax:      "",
		Description: "List albums",
		Fn:          GetAlbumsFn,
	}
	GetSharedAlbums = Command{
		Keyword:     "shared",
		Syntax:      "",
		Description: "List shared albums",
		Fn:          GetSharedAlbumsFn,
	}
)

////////////////////////////////////////////////////////////////////////////////
// COMMAND

func GetAlbumsFn(ctx context.Context, cmd *Command, args []string) error {
	if len(args) != 0 {
		return ErrBadParameter.With("Too many arguments")
	}

	// Authenticate with google
	client, err := cmd.Authenticate(ctx)
	if err != nil {
		return err
	}

	// Retrieve albums
	albums, err := client.AlbumList(ctx, 0, false)
	if err != nil {
		return err
	}

	// Write albums
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Title", "MediaItemsCount", "Writable", "Collaborative"})
	for _, album := range albums {
		table.Append([]string{album.Id, album.Title, album.MediaItemsCount, fmt.Sprint(album.IsWritable), fmt.Sprint(album.ShareInfo.IsCollaborative)})
	}
	table.Render()

	// Return any errors
	return nil
}

func GetSharedAlbumsFn(ctx context.Context, cmd *Command, args []string) error {
	if len(args) != 0 {
		return ErrBadParameter.With("Too many arguments")
	}

	// Authenticate with google
	client, err := cmd.Authenticate(ctx)
	if err != nil {
		return err
	}

	// Retrieve albums
	albums, err := client.SharedAlbumList(ctx, 0, false)
	if err != nil {
		return err
	}

	// Write albums
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Title", "MediaItemsCount", "Writable", "Collaborative"})
	for _, album := range albums {
		table.Append([]string{album.Id, album.Title, album.MediaItemsCount, fmt.Sprint(album.IsWritable), fmt.Sprint(album.ShareInfo.IsCollaborative)})
	}
	table.Render()

	// Return any errors
	return nil
}
