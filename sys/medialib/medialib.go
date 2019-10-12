/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package medialib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"
	tasks "github.com/djthorpe/gopi/util/tasks"
	sqlite "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MediaLib struct {
	Recursive bool
	Media     media.Media
	SQObj     sqlite.Objects
}

type medialib struct {
	log       gopi.Logger
	media     media.Media
	sqobj     sqlite.Objects
	paths     chan string
	recursive bool

	tasks.Tasks
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config MediaLib) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<medialib.Open>{ config=%+v }", config)

	this := new(medialib)
	this.log = logger

	// Set media and sqlite
	if config.Media == nil {
		return nil, gopi.ErrBadParameter
	} else {
		this.media = config.Media
	}
	if config.SQObj == nil {
		return nil, gopi.ErrBadParameter
	} else {
		this.sqobj = config.SQObj
	}

	// Other instance variables
	this.paths = make(chan string)
	this.recursive = config.Recursive

	// Register SQLite tables
	if _, err := this.sqobj.RegisterStruct(&MediaFile{}); err != nil {
		return nil, err
	}
	if _, err := this.sqobj.RegisterStruct(&MediaKey{}); err != nil {
		return nil, err
	}

	// Background scanning process
	this.Start(this.ScanPath)

	// Success
	return this, nil
}

func (this *medialib) Close() error {
	this.log.Debug("<medialib.Close>{ }")

	// stop tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Close channels
	close(this.paths)

	// Release resources
	this.media = nil
	this.sqobj = nil
	this.paths = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *medialib) String() string {
	return fmt.Sprintf("<medialib>{ }")
}

////////////////////////////////////////////////////////////////////////////////
// ADD PATH

func (this *medialib) AddPath(path string) error {
	this.log.Debug2("<medialib>AddPath{ path=%v }", strconv.Quote(path))

	if stat, err := os.Stat(path); os.IsNotExist(err) {
		return err
	} else if err != nil {
		return err
	} else if stat.Size() == 0 {
		return errors.New("Zero-sized file")
	} else if stat.IsDir() {
		this.paths <- path
	} else if stat.Mode().IsRegular() {
		return this.AddFile(path, stat)
	}

	// File not added
	return errors.New("Not a file or folder")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *medialib) WalkPath(path string) error {
	this.log.Debug2("<medialib>WalkPath{ path=%v }", strconv.Quote(path))

	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		// Always skip hidden files and folders
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		// Return any errors
		if err != nil {
			return err
		}
		// Recurse into folders
		if info.IsDir() && this.recursive == false {
			return filepath.SkipDir
		}
		// Skip non-regular files
		if info.Mode().IsRegular() == false {
			return nil
		}
		// Ignore zero-sized files
		if info.Size() == 0 {
			return nil
		}

		// Output path for opening
		return this.AddFile(path, info)
	})
}

func (this *medialib) AddFile(path string, info os.FileInfo) error {
	this.log.Debug2("<medialib>AddFile{ path=%v info=%v }", strconv.Quote(path), info)

	// Read media and extract the metadata from the file
	if mediafile, err := this.media.Open(path); err != nil {
		return err
	} else {
		defer this.media.Destroy(mediafile)
		if id, err := IdForFileInfo(info); err != nil {
			return err
		} else if objs := ObjectsForMediaFile(id, mediafile); objs == nil {
			return gopi.ErrBadParameter
		} else if rowsAffected, err := this.sqobj.Write(sqlite.FLAG_INSERT|sqlite.FLAG_UPDATE, objs...); err != nil {
			return err
		} else {
			fmt.Println(rowsAffected)
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SCAN PATH

func (this *medialib) ScanPath(start chan<- struct{}, stop <-chan struct{}) error {
	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case path := <-this.paths:
			go func() {
				if err := this.WalkPath(path); err != nil {
					this.log.Warn("%v: %v", filepath.Base(path), err)
				}
			}()
		case <-stop:
			break FOR_LOOP
		}
	}
	// Return success
	return nil
}
