/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package media

import (
	"fmt"

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

	// Execute statement without returning the rows
	Do(Statement, ...interface{}) (Result, error)

	// Query to return the rows
	Query(Statement, ...interface{}) (Rows, error)

	// Return sqlite information
	Version() string
	Tables() ([]string, error)
}

// Return rows
type Rows interface {
	Columns() []Column
}

type Column interface {
	Name() string
}

// Result of an insert
type Result struct {
	LastInsertId int64
	RowsAffected uint64
}

// Statement that can be executed
type Statement interface {
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Result) String() string {
	return fmt.Sprintf("<sqlite.Result>{ LastInsertId=%v RowsAffected=%v }", r.LastInsertId, r.RowsAffected)
}
