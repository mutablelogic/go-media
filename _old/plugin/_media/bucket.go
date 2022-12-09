package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Bucket struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type BucketEntry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	ModTime time.Time `json:"modtime"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reBucketName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]+$`)
)

var (
	PathSeparator = string(os.PathSeparator)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewBucket(name, path string) (*Bucket, error) {
	bucket := new(Bucket)

	// Check name is valid
	if !reBucketName.MatchString(name) {
		return nil, ErrBadParameter.With(name)
	}

	// Make path absolute
	if abspath, err := filepath.Abs(path); err != nil {
		return nil, err
	} else {
		path = abspath
	}

	// Check path is valid
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotFound.With(path)
	} else if err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, ErrBadParameter.With(path)
	}

	// Set parameters
	bucket.Name = name
	bucket.Path = path

	// Return success
	return bucket, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (b *Bucket) String() string {
	str := "<bucket"
	str += fmt.Sprintf(" name=%q", b.Name)
	str += fmt.Sprintf(" path=%q", b.Path)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (b *Bucket) FoldersForPath(path string) ([]*BucketEntry, error) {
	return b.entriesForPath(path, func(info fs.FileInfo) bool {
		if strings.HasPrefix(info.Name(), ".") {
			return false
		}
		if !info.IsDir() {
			return false
		}
		return true
	})
}

func (b *Bucket) FilesForPath(path string) ([]*BucketEntry, error) {
	return b.entriesForPath(path, func(info fs.FileInfo) bool {
		if strings.HasPrefix(info.Name(), ".") {
			return false
		}
		if !info.Mode().IsRegular() {
			return false
		}
		return true
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (b *Bucket) entriesForPath(path string, fn func(fs.FileInfo) bool) ([]*BucketEntry, error) {
	path = strings.TrimPrefix(path, PathSeparator)
	abs, err := filepath.Abs(filepath.Join(b.Path, path))
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(abs, b.Path) {
		return nil, ErrBadParameter.With(path)
	}
	files, err := ioutil.ReadDir(abs)
	if err != nil {
		return nil, err
	}

	// Enumerate directories, exclude hidden
	result := make([]*BucketEntry, 0, len(files))
	for _, info := range files {
		if fn(info) {
			rel, err := filepath.Rel(b.Path, filepath.Join(abs, info.Name()))
			if err != nil {
				continue
			}
			result = append(result, &BucketEntry{
				Name:    info.Name(),
				Path:    PathSeparator + strings.TrimPrefix(rel, PathSeparator),
				ModTime: info.ModTime(),
			})
		}
	}

	// return success
	return result, nil
}
