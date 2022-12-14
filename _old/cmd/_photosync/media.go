package main

import (
	"context"
	"os"

	// Packages
	tablewriter "github.com/olekukonko/tablewriter"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GetMedia = Command{
		Keyword:     "media",
		Syntax:      "",
		Description: "List media items",
		Fn:          GetMediaFn,
	}
)

////////////////////////////////////////////////////////////////////////////////
// COMMAND

func GetMediaFn(ctx context.Context, cmd *Command, args []string) error {
	if len(args) != 0 {
		return ErrBadParameter.With("Too many arguments")
	}

	// Authenticate with google
	client, err := cmd.Authenticate(ctx)
	if err != nil {
		return err
	}

	// Retrieve media items
	media, err := client.MediaList(ctx, 0)
	if err != nil {
		return err
	}

	// Display media items
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Description", "mimeType", "filename"})
	for _, media := range media {
		table.Append([]string{media.Id, media.Description, media.MimeType, media.Filename})
	}
	table.Render()

	// Return any errors
	return nil
}

/*

	// Retrieve mediaitems
	if items, err := googlephotos.MediaItemSearch(client); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(-1)
	} else {
		for _, item := range items {
			w, err := os.Create(item.Filename)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(-1)
			}
			defer w.Close()
			if err := googlephotos.DownloadMediaItem(client, w, item, googlephotos.OptWidthHeight(500, 500, true)); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(-1)
			}
		}
	}

*/
