/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package medialib

import (
	media "github.com/djthorpe/gopi-media"
	sqlite "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MediaFile struct {
	Id       int64  `sql:"id,primary"`
	Filename string `sql:"filename"`
	Title_   string `sql:"title"`
	Type_    int64  `sql:"type"`
	sqlite.Object
}

type MediaKey struct {
	Id    int64  `sql:"id,primary"`
	Key   string `sql:"key,primary"`
	Value string `sql:"value"`
	sqlite.Object
}

type MediaArtwork struct {
	Id       int64  `sql:"id,primary"`
	MimeType string `sql:"mime_type"`
	Data     []byte `sql:"data"`
	sqlite.Object
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func ObjectsForMediaFile(id int64, item media.MediaFile) []interface{} {
	if id == 0 || item == nil {
		return nil
	}
	keys := item.Keys()
	objs := make([]interface{}, 1, len(keys)+1)

	// Add Media File
	objs[0] = &MediaFile{id, item.Filename(), item.Title(), int64(item.Type()), sqlite.Object{}}

	// Add in key/value pairs
	for _, k := range keys {
		if value, exists := item.StringForKey(k); exists {
			objs = append(objs, &MediaKey{id, k.String(), value, sqlite.Object{}})
		}
	}

	// Add any artwork
	if data, mime_type := item.ArtworkData(); len(data) > 0 && mime_type != "" {
		objs = append(objs, &MediaArtwork{id, mime_type, data, sqlite.Object{}})
	}

	return objs
}

////////////////////////////////////////////////////////////////////////////////
// MediaItem IMPLEMENTATION

func (this *MediaFile) Keys() []media.MetadataKey {
	return nil
}

func (this *MediaFile) Title() string {
	return this.Title_
}

func (this *MediaFile) Type() media.MediaType {
	return media.MediaType(this.Type_)
}

func (this *MediaFile) StringForKey(media.MetadataKey) (string, bool) {
	return "", false
}
