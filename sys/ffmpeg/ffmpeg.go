/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package ffmpeg

import (
	"fmt"
	"os"
	"strconv"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	media "github.com/djthorpe/gopi-media"
	ff "github.com/djthorpe/gopi-media/ffmpeg"
	errors "github.com/djthorpe/gopi/util/errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
}

type ffmpeg struct {
	log   gopi.Logger
	files []*ffinput
}

type ffinput struct {
	ctx *ff.AVFormatContext
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Config) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<ffmpeg.Open>{ config=%+v }", config)

	this := new(ffmpeg)
	this.log = logger
	this.files = make([]*ffinput, 0)

	// Success
	return this, nil
}

func (this *ffmpeg) Close() error {
	this.log.Debug("<ffmpeg.Close>{ }")

	var err errors.CompoundError
	for _, file := range this.files {
		if file != nil {
			this.log.Debug2("Close: %v", file)
			err.Add(file.Destroy())
		}
	}

	// Return success
	return err.ErrorOrSelf()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *ffmpeg) String() string {
	return fmt.Sprintf("<ffmpeg>{ }")
}

////////////////////////////////////////////////////////////////////////////////
// MEDIA INTERFACE IMPLEMENTATION

func (this *ffmpeg) Open(filename string) (media.MediaFile, error) {
	this.log.Debug2("<ffmpeg.Open>{ filename=%v }", strconv.Quote(filename))

	if stat, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, gopi.ErrNotFound
	} else if err != nil {
		return nil, err
	} else if stat.Mode().IsRegular() == false {
		return nil, gopi.ErrBadParameter
	} else if ctx := ff.NewAVFormatContext(); ctx == nil {
		return nil, gopi.ErrAppError
	} else if err := ctx.OpenInput(filename, nil); err != nil {
		return nil, err
	} else {
		file := &ffinput{ctx: ctx}
		this.files = append(this.files, file)
		return file, nil
	}
}

func (this *ffmpeg) Destroy(file media.MediaFile) error {
	this.log.Debug2("<ffmpeg.Destroy>{ file=%v }", file)
	// TODO
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// MEDIAFILE INTERFACE IMPLEMENTATION

func (this *ffinput) String() string {
	if this.ctx == nil {
		return fmt.Sprintf("<ffinput>{ ctx=nil }")
	} else {
		return fmt.Sprintf("<ffinput>{ filename=%v }", strconv.Quote(this.Filename()))
	}
}

func (this *ffinput) Filename() string {
	if this.ctx == nil {
		return ""
	} else {
		return this.ctx.Filename()
	}
}

func (this *ffinput) Destroy() error {
	if this.ctx == nil {
		return gopi.ErrAppError
	} else {
		this.ctx.Close()
		this.ctx = nil
		return nil
	}
}
