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
	"fmt"
	"strconv"

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

type statement struct {
	p *sql.Stmt
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

func (this *sqlite) Prepare(str string) (sq.Statement, error) {
	this.log.Debug2("<sqlite.Prepare>{ str=%v }", strconv.Quote(str))
	if prepared, err := this.conn.Prepare(str); err != nil {
		return nil, err
	} else {
		return &statement{prepared}, nil
	}
}

func (this *sqlite) Do(st sq.Statement, args ...interface{}) error {
	this.log.Debug2("<sqlite.Do>{ %v }", st)
	if statement_, ok := st.(*statement); ok == false {
		return gopi.ErrBadParameter
	} else if _, err := statement_.p.Exec(args...); err != nil {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// RETURN TABLES

func (this *sqlite) Tables() ([]string, error) {
	if p, err := this.Prepare("SELECT name FROM sqlite_master WHERE type=?"); err != nil {
		return nil, err
	} else if err := this.Do(p, "table"); err != nil {
		return nil, err
	} else {
		return nil, gopi.ErrNotImplemented
	}
}
