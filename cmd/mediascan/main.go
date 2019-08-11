package main

import (
	"fmt"
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"

	// Modules
	_ "github.com/djthorpe/gopi-media/sys/ffmpeg"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	// Get media object
	media := app.ModuleInstance("ffmpeg").(media.Media)

	// Get filenames
	for _, filename := range app.AppFlags.Args() {
		if file, err := media.Open(filename); err != nil {
			app.Logger.Error("%v: %v", filename, err)
		} else {
			fmt.Println(file)
		}
	}

	// Success
	done <- gopi.DONE
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("ffmpeg")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
