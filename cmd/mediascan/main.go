package main

import (
	"fmt"
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"
)

func EventLoop(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	library := app.ModuleInstance("media/library").(media.MediaLibrary)
	messages := library.Subscribe()

	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case evt := <-messages:
			if event, ok := evt.(media.MediaEvent); ok {
				item := event.Item()
				if event.Type() == media.MEDIA_EVENT_FILE_ADDED {
					fmt.Printf("%-10v %-40s %-20s\n", "Added", item.Title(), item.Type())
				} else if event.Type() == media.MEDIA_EVENT_ERROR {
					fmt.Printf("%-10v %-20s %s\n", "Error", event.Error(), event.Path())
				} else if event.Type() == media.MEDIA_EVENT_SCAN_START {
					fmt.Printf("%-10v %s\n", "Started", event.Path())
				} else if event.Type() == media.MEDIA_EVENT_SCAN_END {
					fmt.Printf("%-10v %s\n", "Ended", event.Path())
				}
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	// End of routine
	library.Unsubscribe(messages)
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	library := app.ModuleInstance("media/library").(media.MediaLibrary)

	// Get paths and add to library
	for _, path := range app.AppFlags.Args() {
		if stat, err := os.Stat(path); os.IsNotExist(err) {
			app.Logger.Warn("%v: Not Found", path)
		} else if err != nil {
			app.Logger.Warn("%v: %v", path, err)
		} else if stat.IsDir() || stat.Mode().IsRegular() {
			library.AddPath(path)
		} else {
			app.Logger.Warn("%v: Not added", path)
		}
	}

	// Wait for CTRL+C
	app.Logger.Info("Waiting for CTRL+C")
	app.WaitForSignal()

	// Success
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("media/library")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, EventLoop))
}
