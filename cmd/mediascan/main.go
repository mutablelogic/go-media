package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"

	// Modules
	_ "github.com/djthorpe/gopi-media/sys/ffmpeg"
	_ "github.com/djthorpe/gopi-media/sys/sqlite"
	_ "github.com/djthorpe/gopi/sys/logger"
)

var (
	paths = make(chan string)
	files = make(chan string)
)

func WalkFiles(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	// Get media object, start processing
	ffmpeg := app.ModuleInstance("ffmpeg").(media.Media)
	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case filename := <-files:
			// Ignore files by guessing their type
			if ffmpeg.TypeFor(filename) == media.MEDIA_TYPE_NONE {
				app.Logger.Warn("Ignoring %v", filename)
			} else if file, err := ffmpeg.Open(filename); err != nil {
				app.Logger.Error("%v: %v", filename, err)
			} else {
				fmt.Println(file)
				ffmpeg.Destroy(file)
			}
		case <-stop:
			break FOR_LOOP
		}
	}
	return nil
}

func WalkPaths(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	recursive := true
	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case path := <-paths:
			if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				// Always skip hidden files
				if strings.HasPrefix(info.Name(), ".") {
					return nil
				}
				// Return any errors
				if err != nil {
					return err
				}
				// Recurse into folders
				if info.IsDir() && recursive == false {
					return filepath.SkipDir
				}
				// Skip non-regular files
				if info.Mode().IsRegular() == false {
					return nil
				}

				// Ignore zero-sized files
				if info.Size() == 0 {
					return nil
				}

				// Output path for opening
				files <- path

				// Return success
				return nil
			}); err != nil {
				app.Logger.Error("%v: %v", path, err)
			}
		case <-stop:
			break FOR_LOOP
		}
	}
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	// Get paths
	for _, path := range app.AppFlags.Args() {
		if stat, err := os.Stat(path); os.IsNotExist(err) {
			app.Logger.Error("%v: Not Found", path)
		} else if err != nil {
			app.Logger.Error("%v: %v", path, err)
		} else if stat.IsDir() {
			paths <- path
		} else if stat.Mode().IsRegular() {
			files <- path
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
	config := gopi.NewAppConfig("ffmpeg", "sqlite")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, WalkPaths, WalkFiles))
}
