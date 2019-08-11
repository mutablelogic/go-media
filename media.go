/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package media

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Media interface {
	gopi.Driver

	// Open and close media files
	Open(filename string) (MediaFile, error)
	Destroy(MediaFile) error
}

type MediaFile interface {
	Filename() string
}
