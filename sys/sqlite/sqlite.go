/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package sqlite

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/gopi-media"
	errors "github.com/djthorpe/gopi/util/errors"

	// Anonymous
	_ "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Path string
}

type sqlite struct {
	log  gopi.Logger
	conn *sql.DB
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Config) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<sqlite.Open>{ config=%+v }", config)

	this := new(sqlite)
	this.log = logger

	if db, err := sql.Open("sqlite3", config.Path); err != nil {
		return nil, err
	} else {
		this.conn = db
	}

	// Success
	return this, nil
}

func (this *sqlite) Close() error {
	this.log.Debug("<sqlite.Close>{ conn=%v }", this.conn)

	var err errors.CompoundError

	err.Add(this.conn.Close())

	// Return success
	return err.ErrorOrSelf()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *sqlite) String() string {
	return fmt.Sprintf("<sqlite>{ conn=%v }", this.conn)
}

////////////////////////////////////////////////////////////////////////////////
// STATEMENT PREPARE AND EXECUTE

func (this *sqlite) Prepare(statement string) (sq.Statement, error) {
	if _, err := this.conn.Prepare(statement); err != nil {
		return nil, err
	} else {
		return nil, gopi.ErrNotImplemented
	}
}

func (this *sqlite) Do(statement sq.Statement) error {
	this.log.Debug2("<sqlite.Do>{ %v }", statement.SQL())
	if s, err := this.conn.Prepare(statement.SQL()); err != nil {
		return err
	} else if _, err := s.Exec([]driver.Value{}); err != nil {
		return err
	} else {
		return nil
	}
	return gopi.ErrNotImplemented
}
