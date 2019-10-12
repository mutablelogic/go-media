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
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	files map[*ffinput]bool
}

type ffinput struct {
	log  gopi.Logger
	ctx  *ff.AVFormatContext
	keys map[media.MetadataKey]string

	sync.Mutex
}

type ffstream struct {
	ctx *ff.AVStream
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Config) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<ffmpeg.Open>{ config=%+v }", config)

	// Init ffmpeg
	ff.AVFormatInit()

	this := new(ffmpeg)
	this.log = logger
	this.files = make(map[*ffinput]bool, 0)

	// Success
	return this, nil
}

func (this *ffmpeg) Close() error {
	this.log.Debug("<ffmpeg.Close>{ }")

	var err errors.CompoundError
	for file, _ := range this.files {
		err.Add(file.Destroy())
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

	if file, err := NewInput(filename, this.log); err != nil {
		return nil, err
	} else {
		// Append file to list of opened files
		this.files[file] = true
		return file, nil
	}
}

func (this *ffmpeg) Destroy(file media.MediaFile) error {
	this.log.Debug2("<ffmpeg.Destroy>{ file=%v }", file)

	// If in list of files, then remove
	if file_, ok := file.(*ffinput); ok == false {
		return gopi.ErrBadParameter
	} else if _, exists := this.files[file_]; exists == false {
		return gopi.ErrNotFound
	} else if err := file_.Destroy(); err != nil {
		return err
	} else {
		delete(this.files, file_)
		return nil
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
		// ctx is freed on error, so no need to free here
		return nil, err
	} else if err := ctx.ReadHeader(); err != nil {
		ctx.CloseInput()
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
	this.Lock()
	defer this.Unlock()

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
		return fmt.Sprintf("<ffinput>{ filename=%v type=%v metadata={%v} streams=%v }", strconv.Quote(this.Filename()), this.Type(), strings.TrimSpace(metadata), this.Streams())
	}
}

func (this *ffinput) Filename() string {
	this.log.Debug2("<ffinput.Filename>{}")
	this.Lock()
	defer this.Unlock()
	if this.ctx == nil {
		return ""
	} else {
		return this.ctx.Filename()
	}
}

func (this *ffinput) Streams() []media.MediaStream {
	this.log.Debug2("<ffinput.Streams>{}")
	this.Lock()
	defer this.Unlock()
	if this.ctx == nil {
		return nil
	}
	streams := make([]media.MediaStream, this.ctx.NumStreams())
	for i, stream := range this.ctx.Streams() {
		streams[i] = NewStream(stream)
	}
	return streams
}

////////////////////////////////////////////////////////////////////////////////
// MEDIAITEM INTERFACE IMPLEMENTATION

func (this *ffinput) Keys() []media.MetadataKey {
	this.log.Debug2("<ffinput.Keys>{}")
	this.Lock()
	defer this.Unlock()
	if this.ctx == nil {
		return nil
	}

	keys := make([]media.MetadataKey, 0, len(this.keys))
	for k := range this.keys {
		keys = append(keys, k)
	}
	return keys
}

func (this *ffinput) StringForKey(key media.MetadataKey) (string, bool) {
	this.log.Debug2("<ffinput.StringForKey>{ key=%v }", key)
	this.Lock()
	defer this.Unlock()
	if this.ctx == nil {
		return "", false
	}
	if value, exists := this.keys[key]; exists {
		return value, true
	} else {
		return "", false
	}
}

func (this *ffinput) Title() string {
	if title, exists := this.StringForKey(media.METADATA_KEY_TITLE); exists {
		// Obtain title from media key
		return title
	} else {
		// Obtain title from filename
		base := filepath.Base(this.Filename())
		if ext := filepath.Ext(base); ext != "" {
			base = strings.TrimSuffix(base, ext)
		}
		return base
	}
}

func (this *ffinput) Type() media.MediaType {
	if this.ctx == nil {
		return media.MEDIA_TYPE_NONE
	}

	media_type := media.MEDIA_TYPE_FILE

	// Add in all the stream types
	for _, stream := range this.Streams() {
		media_type |= stream.Type()
	}

	// If there is a key METADATA_KEY_ALBUM then this is an album song
	if _, exists := this.keys[media.METADATA_KEY_ALBUM]; exists {
		media_type |= media.MEDIA_TYPE_ALBUM
	}

	// If there is a key METADATA_KEY_SHOW then this is a TV episode
	if _, exists := this.keys[media.METADATA_KEY_SHOW]; exists {
		media_type |= media.MEDIA_TYPE_TVEPISODE
	}

	// Compilations
	if value, exists := this.keys[media.METADATA_KEY_COMPILATION]; exists {
		if int_value, err := strconv.ParseInt(value, 10, 32); err == nil && int_value != 0 {
			media_type |= media.MEDIA_TYPE_COMPILATION
		}
	}

	// If the extension is .m4r then ringtone
	if value, exists := this.keys[media.METADATA_KEY_EXTENSION]; exists {
		if strings.ToLower(value) == ".m4r" {
			media_type |= media.MEDIA_TYPE_RINGTONE
		}
	}

	// Return combined media type
	return media_type
}

// ArtworkData returns raw image data for artwork, or
// nil if no artwork associated with the file
func (this *ffinput) ArtworkData() ([]byte, string) {
	this.log.Debug2("<ffinput.ArtworkData>{}")
	if this.ctx == nil {
		return nil, ""
	}
	for _, stream := range this.Streams() {
		if stream.Type() == media.MEDIA_TYPE_ARTWORK {
			if data := stream.(*ffstream).AttachedPic(); len(data) > 0 {
				return data, http.DetectContentType(data)
			}
		}
	}
	// No artwork associated
	return nil, ""
}

////////////////////////////////////////////////////////////////////////////////
// MEDIASTREAM INTERFACE IMPLEMENTATION

func NewStream(ctx *ff.AVStream) media.MediaStream {
	if ctx == nil {
		return nil
	}
	return &ffstream{ctx}
}

func (this *ffstream) Type() media.MediaType {
	flags := media.MEDIA_TYPE_NONE

	// Codec flags
	switch this.ctx.Codec().Type() {
	case ff.AV_MEDIA_TYPE_VIDEO:
		if this.ctx.Codec().BitRate() > 0 {
			flags |= media.MEDIA_TYPE_VIDEO
		}
	case ff.AV_MEDIA_TYPE_AUDIO:
		flags |= media.MEDIA_TYPE_AUDIO
	case ff.AV_MEDIA_TYPE_UNKNOWN, ff.AV_MEDIA_TYPE_DATA:
		flags |= media.MEDIA_TYPE_DATA
	case ff.AV_MEDIA_TYPE_ATTACHMENT:
		flags |= media.MEDIA_TYPE_ATTACHMENT
	}

	// Disposition flags
	if this.ctx.Disposition()&ff.AV_DISPOSITION_ATTACHED_PIC != 0 {
		flags |= media.MEDIA_TYPE_ARTWORK
	}
	if this.ctx.Disposition()&ff.AV_DISPOSITION_CAPTIONS != 0 {
		flags |= media.MEDIA_TYPE_CAPTIONS
	}

	// Return flags
	return flags
}

func (this *ffstream) AttachedPic() []byte {
	if this.ctx.Disposition()&ff.AV_DISPOSITION_ATTACHED_PIC != 0 {
		return this.ctx.AttachedPic().Data()
	} else {
		return nil
	}
}

func (this *ffstream) String() string {
	if this.ctx == nil {
		return fmt.Sprintf("<ffstream>{ nil }")
	} else {
		return fmt.Sprintf("<ffstream>{ type=%v %v }", this.Type(), this.ctx.String())
	}
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
