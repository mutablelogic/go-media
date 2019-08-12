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
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	log  gopi.Logger
	ctx  *ff.AVFormatContext
	keys map[media.MetadataKey]string
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Config) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<ffmpeg.Open>{ config=%+v }", config)

	// Init ffmpeg
	ff.AVFormatInit()

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

	// Release resources
	this.files = nil

	// Deallocate for AVFormat
	ff.AVFormatDeinit()

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
	} else if file, err := NewInput(filename, this.log); err != nil {
		return nil, err
	} else {
		this.files = append(this.files, file)
		return file, nil
	}
}

func (this *ffmpeg) Destroy(file media.MediaFile) error {
	this.log.Debug2("<ffmpeg.Destroy>{ file=%v }", file)
	// TODO
	return gopi.ErrNotImplemented
}

func (this *ffmpeg) TypeFor(filename string) media.MediaType {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4", ".m4v", ".mov", ".m2v", ".vob":
		return media.MEDIA_TYPE_MOVIE
	case ".mp3", ".aac", ".m4a":
		return media.MEDIA_TYPE_MUSIC
	case ".m4r":
		return media.MEDIA_TYPE_RINGTONE
	default:
		return media.MEDIA_TYPE_NONE
	}
}

////////////////////////////////////////////////////////////////////////////////
// MEDIAFILE INTERFACE IMPLEMENTATION

func NewInput(filename string, log gopi.Logger) (*ffinput, error) {
	if stat, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, gopi.ErrNotFound
	} else if err != nil {
		return nil, err
	} else if ctx := ff.NewAVFormatContext(); ctx == nil {
		return nil, gopi.ErrAppError
	} else if err := ctx.OpenInput(filename, nil); err != nil {
		ctx.Free()
		return nil, err
	} else {
		dict := ctx.Metadata()
		this := new(ffinput)
		this.log = log
		this.ctx = ctx
		this.keys = make(map[media.MetadataKey]string, dict.Count()+10)

		// Set the file attributes
		this.keys[media.METADATA_KEY_FILENAME] = filename
		this.keys[media.METADATA_KEY_FILESIZE] = fmt.Sprint(stat.Size())
		this.keys[media.METADATA_KEY_EXTENSION] = filepath.Ext(filename)
		this.keys[media.METADATA_KEY_MODIFIED] = stat.ModTime().Format(time.RFC3339)

		// Read the metadata
		for _, entry := range dict.Entries() {
			entry_key := entry.Key()
			if strings.HasPrefix(entry_key, "iTun") || entry_key == "Encoding Params" {
				// We ignore any iTunes-specific metadata
				this.log.Debug2("Ignoring metadata entry: %v", entry)
			} else if key := MetadataKeyFor(entry_key); key != media.METADATA_KEY_NONE {
				this.keys[key] = entry.Value()
			} else {
				this.log.Warn("Ignoring metadata entry: %v", entry)
			}
		}

		return this, nil
	}
}

func (this *ffinput) Destroy() error {
	this.log.Debug2("<ffinput.Destroy>{ ctx=%v }", this.ctx)

	if this.ctx == nil {
		// Do nothing - already closed
		this.keys = nil
		return nil
	} else {
		this.ctx.CloseInput()
		this.ctx = nil
		this.keys = nil
		return nil
	}
}

func (this *ffinput) String() string {
	if this.ctx == nil {
		return fmt.Sprintf("<ffinput>{ ctx=nil }")
	} else {
		metadata := ""
		for k, v := range this.keys {
			metadata_key := strings.ToLower(strings.TrimPrefix(fmt.Sprint(k), "METADATA_KEY_"))
			metadata_value := strconv.Quote(v)
			metadata += fmt.Sprintf("%v=%v ", metadata_key, metadata_value)
		}
		return fmt.Sprintf("<ffinput>{ filename=%v metadata={%v} }", strconv.Quote(this.Filename()), strings.TrimSpace(metadata))
	}
}

func (this *ffinput) Filename() string {
	if this.ctx == nil {
		return ""
	} else {
		return this.ctx.Filename()
	}
}

////////////////////////////////////////////////////////////////////////////////
// MEDIAITEM INTERFACE IMPLEMENTATION

func (this *ffinput) Keys() []media.MetadataKey {
	if this.ctx == nil {
		return nil
	} else {
		return nil
	}
}

func (this *ffinput) StringForKey(media.MetadataKey) string {
	this.log.Warn("TODO: StringForKey")
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// CONVERT FFMPEG KEYS

func MetadataKeyFor(key string) media.MetadataKey {
	switch key {
	case "major_brand":
		return media.METADATA_KEY_BRAND_MAJOR
	case "compatible_brands":
		return media.METADATA_KEY_BRAND_COMPATIBLE
	case "creation_time":
		return media.METADATA_KEY_CREATED
	case "encoder":
		return media.METADATA_KEY_ENCODER
	case "album":
		return media.METADATA_KEY_ALBUM
	case "album_artist":
		return media.METADATA_KEY_ALBUM_ARTIST
	case "artist":
		return media.METADATA_KEY_ARTIST
	case "comment":
		return media.METADATA_KEY_COMMENT
	case "composer":
		return media.METADATA_KEY_COMPOSER
	case "copyright":
		return media.METADATA_KEY_COPYRIGHT
	case "date":
		return media.METADATA_KEY_YEAR
	case "disc":
		return media.METADATA_KEY_DISC
	case "encoded_by":
		return media.METADATA_KEY_ENCODED_BY
	case "filename":
		return media.METADATA_KEY_FILENAME
	case "genre":
		return media.METADATA_KEY_GENRE
	case "language":
		return media.METADATA_KEY_LANGUAGE
	case "performer":
		return media.METADATA_KEY_PERFORMER
	case "publisher":
		return media.METADATA_KEY_PUBLISHER
	case "service_name":
		return media.METADATA_KEY_SERVICE_NAME
	case "service_provider":
		return media.METADATA_KEY_SERVICE_PROVIDER
	case "title":
		return media.METADATA_KEY_TITLE
	case "track":
		return media.METADATA_KEY_TRACK
	case "major_version":
		return media.METADATA_KEY_VERSION_MAJOR
	case "minor_version":
		return media.METADATA_KEY_VERSION_MINOR
	case "show":
		return media.METADATA_KEY_SHOW
	case "season_number":
		return media.METADATA_KEY_SEASON
	case "episode_sort":
		return media.METADATA_KEY_EPISODE_SORT
	case "episode_id":
		return media.METADATA_KEY_EPISODE_ID
	case "compilation":
		return media.METADATA_KEY_COMPILATION
	case "gapless_playback":
		return media.METADATA_KEY_GAPLESS_PLAYBACK
	case "account_id":
		return media.METADATA_KEY_ACCOUNT_ID
	case "description":
		return media.METADATA_KEY_DESCRIPTION
	case "media_type":
		return media.METADATA_KEY_MEDIA_TYPE
	case "purchase_date":
		return media.METADATA_KEY_PURCHASED
	case "sort_album":
		return media.METADATA_KEY_ALBUM_SORT
	case "sort_artist":
		return media.METADATA_KEY_ARTIST_SORT
	case "sort_name":
		return media.METADATA_KEY_TITLE_SORT
	case "synopsis":
		return media.METADATA_KEY_SYNOPSIS
	case "grouping":
		return media.METADATA_KEY_GROUPING
	default:
		return media.METADATA_KEY_NONE
	}
}
