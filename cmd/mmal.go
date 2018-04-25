package main

import (
	"fmt"
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/mmal"

	// Modules
	_ "github.com/djthorpe/gopi/sys/hw/rpi"
	_ "github.com/djthorpe/gopi/sys/logger"

	// MMAL
	rpi_mmal "github.com/djthorpe/mmal/sys/hw/mmal"
)

/////////////////////////////////////////////////////////////////////
// MAIN

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	name, _ := app.AppFlags.GetString("component")

	if component_, err := gopi.Open(rpi_mmal.Component{
		Hardware: app.Hardware,
		Name:     name,
	}, app.Logger); err != nil {
		return err
	} else {
		component := component_.(mmal.Component)
		defer component.Close()

		// Enable component
		if err := component.SetEnabled(true); err != nil {
			return err
		}

		fmt.Println("component", component)

		fmt.Println("control", component.Control())
		component.Control().SupportedEncodings()
		fmt.Println("")

		for _, port := range component.Input() {
			fmt.Println("input", port)
			if encodings, err := port.SupportedEncodings(); err != nil {
				return err
			} else {
				fmt.Println(" encodings=", encodings)
			}
		}
		for _, port := range component.Output() {
			fmt.Println("output", port)
			if encodings, err := port.SupportedEncodings(); err != nil {
				return err
			} else {
				fmt.Println(" encodings=", encodings)
			}
		}
	}

	// return success
	done <- gopi.DONE
	return nil
}

func main() {
	config := gopi.NewAppConfig("hw")
	config.AppFlags.FlagString("component", mmal.MMAL_COMPONENT_DEFAULT_VIDEO_DECODER, "MMAL Compnent")

	os.Exit(gopi.CommandLineTool(config, Main))
}
