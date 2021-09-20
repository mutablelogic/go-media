package main

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-media"

	// Packages
	media "github.com/djthorpe/go-media/pkg/media"
	multierror "github.com/hashicorp/go-multierror"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type File struct {
	Path  string
	Size  int64
	Meta  map[MediaKey]interface{}
	Flags MediaFlag
}

type Metadata struct {
	*media.Manager
	files []*File
	keys  map[MediaKey]bool
}

type MetaIterator struct {
	n     int
	files []*File
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GetMetadata = Command{
		Keyword:     "metadata",
		Syntax:      "<file>...",
		Description: "Print metadata for one or more files",
		Fn:          GetMetadataFn,
	}
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMetadata(cfg media.Config, errs chan<- error) (*Metadata, error) {
	meta := new(Metadata)
	mgr, err := media.NewManagerWithConfig(cfg, errs)
	if err != nil {
		return nil, err
	}
	meta.Manager = mgr
	meta.keys = make(map[MediaKey]bool)
	return meta, nil
}

func NewFile(path string, info fs.FileInfo, media *media.MediaInput) *File {
	file := new(File)
	file.Path = path
	file.Size = info.Size()
	file.Meta = make(map[MediaKey]interface{})
	for _, key := range media.Metadata().Keys() {
		file.Meta[key] = media.Metadata().Value(key)
	}
	file.Flags = media.Flags() &^ (MEDIA_FLAG_DECODER | MEDIA_FLAG_FILE)
	return file
}

func (meta *Metadata) Close() error {
	var result error

	if err := meta.Manager.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Release resources
	meta.files = nil
	meta.keys = nil
	meta.Manager = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// COMMAND

func GetMetadataFn(ctx context.Context, cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrBadParameter.With("Missing filename argument")
	}

	// Make metadata object
	meta, err := NewMetadata(media.DefaultConfig, cmd.Errs)
	if err != nil {
		return err
	}
	defer meta.Close()

	// Make file walker
	walker := NewWalkFS(meta.Visit)

	// Set up file extensions we will include
	for _, format := range meta.Formats() {
		for _, ext := range strings.Split(format.Ext(), ",") {
			if ext != "" {
				walker.IncludeExt(ext)
			}
		}
	}

	// Routine to print number of files parsed
	var wg sync.WaitGroup
	wg.Add(1)
	child, cancel := context.WithCancel(ctx)
	go func() {
		defer wg.Done()
		timer := time.NewTicker(time.Second * 5)
		defer timer.Stop()
		for {
			select {
			case <-child.Done():
				return
			case <-timer.C:
				fmt.Printf("   ...%d files parsed\n", walker.Count())
			}
		}
	}()

	// Walk filesystem, appending files
	result := walker.WalkArgs(ctx, args...)

	// End reporting
	cancel()
	wg.Wait()

	// Print metadata
	iter := meta.Iterator()
	for {
		file := iter.Next()
		if file == nil {
			break
		}
		fmt.Println(file)
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (file *File) String() string {
	str := "<file"
	str += fmt.Sprintf(" name=%q", filepath.Base(file.Path))
	str += fmt.Sprint(" size=", file.Size)
	for k, v := range file.Meta {
		if v_, ok := v.(string); ok {
			str += fmt.Sprintf(" %v=%q", k, v_)
		} else {
			str += fmt.Sprintf(" %v=%v", k, v)
		}
	}
	if file.Flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", file.Flags)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (meta *Metadata) Visit(ctx context.Context, path string, info fs.FileInfo) error {
	f, err := meta.Manager.OpenFile(path)
	if err != nil {
		return err
	}
	defer meta.Manager.Release(f)

	// Append file
	meta.append(NewFile(path, info, f))

	// Return success
	return nil
}

func (meta *Metadata) Iterator() *MetaIterator {
	return &MetaIterator{0, meta.files}
}

func (iter *MetaIterator) Next() *File {
	if iter.n >= len(iter.files) {
		return nil
	} else {
		iter.n++
		return iter.files[iter.n-1]
	}
}

func (file *File) Keys() []MediaKey {
	result := make([]MediaKey, len(file.Meta))
	for k := range file.Meta {
		result = append(result, k)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (meta *Metadata) append(file *File) {
	// Add media keys
	for _, key := range file.Keys() {
		meta.keys[key] = true
	}
	// Append file
	meta.files = append(meta.files, file)
}
