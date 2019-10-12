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
	Id       int64
	Filename string
	Title    string
	Type     int64
	sqlite.Object
}

type MediaKey struct {
	Id    int64
	Key   string
	Value string
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
	objs[0] = &MediaFile{id, item.Filename(), item.Title(), int64(item.Type()), sqlite.Object{}}
	for _, k := range keys {
		objs = append(objs, &MediaKey{id, k.String(), item.StringForKey(k), sqlite.Object{}})
	}
	return objs
}
