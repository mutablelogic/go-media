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
// INTERFACES

// Database connection
type Connection interface {
	gopi.Driver

	// Prepare statement
	Prepare(string) (Statement, error)

	// Execute statement, no resultset
	Do(Statement, ...interface{}) error
}

// Statement that can be executed
type Statement interface {
}
