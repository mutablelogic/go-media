package main

import (
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"
)

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
	os.Exit(gopi.CommandLineTool2(config, Main))
}
