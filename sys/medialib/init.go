/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package medialib

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"
	sqlite "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	gopi.RegisterModule(gopi.Module{
		Name:     "media/library",
		Requires: []string{"media/ffmpeg", "db/sqobj"},
		Type:     gopi.MODULE_TYPE_OTHER,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagBool("medialib.recursive", true, "Scan folders recursively")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			recursive, _ := app.AppFlags.GetBool("medialib.recursive")
			return gopi.Open(MediaLib{
				Recursive: recursive,
				Media:     app.ModuleInstance("media/ffmpeg").(media.Media),
				SQObj:     app.ModuleInstance("db/sqobj").(sqlite.Objects),
			}, app.Logger)
		},
	})
}
