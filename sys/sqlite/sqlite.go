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
	driver "github.com/mattn/go-sqlite3"
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

type resultset struct {
	r *sql.Rows
	c []string
	t []*sql.ColumnType
}

type column struct {
	name string
	pos  uint
	t    *sql.ColumnType
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
	return fmt.Sprintf("<sqlite>{ version=%v }", strconv.Quote(this.Version()))
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

func (this *sqlite) Do(st sq.Statement, args ...interface{}) (sq.Result, error) {
	this.log.Debug2("<sqlite.Do>{ %v }", st)
	if statement_, ok := st.(*statement); ok == false {
		return sq.Result{}, gopi.ErrBadParameter
	} else if r, err := statement_.p.Exec(args...); err != nil {
		return sq.Result{}, err
	} else if lastInsertID, err := r.LastInsertId(); err != nil {
		return sq.Result{}, err
	} else if rowsAffected, err := r.RowsAffected(); err != nil {
		return sq.Result{}, err
	} else {
		return sq.Result{lastInsertID, uint64(rowsAffected)}, nil
	}
}

func (this *sqlite) Query(st sq.Statement, args ...interface{}) (sq.Rows, error) {
	this.log.Debug2("<sqlite.Query>{ %v }", st)
	if statement_, ok := st.(*statement); ok == false {
		return nil, gopi.ErrBadParameter
	} else if rows, err := statement_.p.Query(args...); err != nil {
		return nil, err
	} else if columns, err := rows.Columns(); err != nil {
		return nil, err
	} else if types, err := rows.ColumnTypes(); err != nil {
		return nil, err
	} else if len(columns) != len(types) {
		return nil, gopi.ErrAppError
	} else {
		return &resultset{
			r: rows,
			c: columns,
			t: types,
		}, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// RESULTSET

func (this *resultset) Columns() []sq.Column {
	c := make([]sq.Column, len(this.c))
	for i, name := range this.c {
		c[i] = &column{name, uint(i), this.t[i]}
	}
	return c
}

////////////////////////////////////////////////////////////////////////////////
// COLUMN

func (this *column) Name() string {
	return this.name
}

////////////////////////////////////////////////////////////////////////////////
// RETURN METADATA

func (this *sqlite) Version() string {
	version, _, _ := driver.Version()
	return version
}

func (this *sqlite) Tables() ([]string, error) {
	if p, err := this.Prepare("SELECT * FROM sqlite_master WHERE type=?"); err != nil {
		return nil, err
	} else if rows, err := this.Query(p, "table"); err != nil {
		return nil, err
	} else {
		this.log.Info("r=%v", rows)
		return nil, gopi.ErrNotImplemented
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *resultset) String() string {
	return fmt.Sprintf("<sqlite.Rows>{ columns=%v }", r.Columns())
}

func (t *column) String() string {
	return fmt.Sprintf("<sqlite.Column>{ name=%v pos=%v }", strconv.Quote(t.name), t.pos)
}
