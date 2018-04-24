package main

import (
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/mmal"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/rpi"
	_ "github.com/djthorpe/gopi/sys/logger"

	// MMAL
	"github.com/djthorpe/mmal/sys/hw/rpi"
)

/////////////////////////////////////////////////////////////////////
// MAIN

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if component_, err := gopi.Open(rpi.MMALComponent{
		Hardware: app.Hardware,
		Name:     mmal.MMAL_COMPONENT_DEFAULT_CAMERA_INFO,
	}, app.Logger); err != nil {
		return err
	} else {
		component := component_.(mmal.Component)
		defer component.Close()

		// Enable component
		if err := component.SetEnabled(true); err != nil {
			return err
		}

		app.Logger.Info("component=%v", component)
	}

	// return success
	done <- gopi.DONE
	return nil
}

func main() {
	config := gopi.NewAppConfig("hw")
	os.Exit(gopi.CommandLineTool(config, Main))
}
