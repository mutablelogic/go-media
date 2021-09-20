/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sqlite

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	gopi.RegisterModule(gopi.Module{
		Name: "sqlite",
		Type: gopi.MODULE_TYPE_OTHER,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("sqlite.path", ":memory:", "Path to database")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			path, _ := app.AppFlags.GetString("sqlite.path")
			return gopi.Open(Config{
				Path: path,
			}, app.Logger)
		},
	})
}
