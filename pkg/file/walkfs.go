package file

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// VisitFunc is a function which is called for each file and folder visited.
// The function should return an error if the walk should be
// terminated.
type VisitFunc func(ctx context.Context, abspath string, relpath string, info fs.FileInfo) error

type WalkFS struct {
	sync.Mutex
	inext   map[string]bool // extensions to include
	exext   map[string]bool // extensions to exclude
	expath  map[string]bool // paths to exclude
	exname  map[string]bool // names to exclude
	count   int
	visitfn VisitFunc
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	pathSeparator = string(os.PathSeparator)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new walkfs with a given visitor function, which is used for
// touching each visited file and folder
func NewWalker(fn VisitFunc) *WalkFS {
	walkfs := new(WalkFS)
	walkfs.inext = make(map[string]bool)
	walkfs.exext = make(map[string]bool)
	walkfs.expath = make(map[string]bool)
	walkfs.exname = make(map[string]bool)
	walkfs.visitfn = fn
	return walkfs
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Count the number of files and folders visited
func (walkfs *WalkFS) Count() int {
	return walkfs.count
}

// Include adds a file extension inclusion to the indexer.
// Path exclusions are case-sensitive, file extension exclusions are not.
// If no inclusions are added, all files are visited
func (walkfs *WalkFS) Include(ext string) error {
	ext = strings.TrimSpace(ext)
	ext = strings.ToUpper("." + strings.TrimPrefix(ext, "."))
	if ext != "." {
		walkfs.inext[ext] = true
	} else {
		return ErrBadParameter.Withf("invalid inclusion: %q", ext)
	}

	// Return success
	return nil
}

// Exclude adds a path or file extension exclusion to the indexer.
// If it begins with a '.' then a file extension exlusion is added,
// If it begins with a '/' then a path extension exclusion is added.
// Path and name exclusions are case-sensitive, file extension exclusions are not.
func (walkfs *WalkFS) Exclude(v string) error {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, ".") && v != "." {
		v = strings.ToUpper(v)
		walkfs.exext[v] = true
	} else if strings.HasPrefix(v, pathSeparator) && v != pathSeparator {
		v = pathSeparator + strings.Trim(v, pathSeparator)
		walkfs.expath[v] = true
	} else if !strings.Contains(v, pathSeparator) && v != "" {
		walkfs.exname[v] = true
	} else {
		return ErrBadParameter.Withf("invalid exclusion: %q", v)
	}

	// Return success
	return nil
}

// Walk will walk a file or folder and visit the function for each
func (walkfs *WalkFS) Walk(ctx context.Context, path string) error {
	walkfs.Mutex.Lock()
	defer walkfs.Mutex.Unlock()

	walkfs.count = 0
	if abspath, err := filepath.Abs(path); err != nil {
		return ErrNotFound.With(path)
	} else {
		path = abspath
	}
	if stat, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return ErrNotFound.With(path)
	} else if err != nil {
		return err
	} else if stat.IsDir() {
		if err := walkfs.walk(ctx, path); err != nil {
			return err
		}
	} else if stat.Mode().IsRegular() {
		if err := walkfs.visit(ctx, "", path, stat); err != nil {
			return err
		}
	} else {
		return ErrNotFound.With(path)
	}

	// Return success
	return nil
}

// ShouldVisit returns true if a path or file should be visited based
// on exclusions or else returns false
func (walkfs *WalkFS) ShouldVisit(relpath string, info fs.FileInfo) bool {
	if !walkfs.shouldVisit(info) {
		return false
	}
	if walkfs.shouldExcludePath(relpath) {
		return false
	}
	if info.Mode().IsRegular() && walkfs.shouldExcludeFile(info) {
		return false
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (walkfs *WalkFS) walk(ctx context.Context, abspath string) error {
	// Walk filesystem
	var result error
	err := filepath.WalkDir(abspath, func(path string, file fs.DirEntry, err error) error {
		// Bail out on context error
		if ctx.Err() != nil {
			return ctx.Err()
		} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// Ignore hidden files and folders
		if strings.HasPrefix(file.Name(), ".") {
			if file.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Process files which can be read
		if relpath, err := filepath.Rel(abspath, path); err == nil {
			if info, err := file.Info(); err == nil {
				if err := walkfs.visit(ctx, abspath, relpath, info); err != nil {
					if errors.Is(filepath.SkipDir, err) {
						return filepath.SkipDir
					} else {
						result = errors.Join(result, err)
					}
				}
				return nil
			}
		}
		// Return any context error
		return ctx.Err()
	})

	// Return errors unless they are cancel/timeout
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return result
	} else if err != nil {
		return errors.Join(result, err)
	} else {
		return nil
	}
}

func (walkfs *WalkFS) visit(ctx context.Context, abspath, relpath string, info fs.FileInfo) error {
	walkfs.count++
	if !walkfs.ShouldVisit(relpath, info) {
		if info.IsDir() {
			return filepath.SkipDir
		} else {
			return nil
		}
	} else if walkfs.visitfn != nil {
		return walkfs.visitfn(ctx, abspath, relpath, info)
	} else {
		return nil
	}
}

// shouldVisit returns true if the given directory entry should be visited
func (walkfs *WalkFS) shouldVisit(info fs.FileInfo) bool {
	// Include all files if no inclusions are specified
	if len(walkfs.inext) == 0 {
		return true
	}
	// Should visit all folders
	if info.Mode().IsDir() {
		return true
	}
	// Ignore anything which isn't a regular file
	if !info.Mode().IsRegular() {
		return false
	}
	ext := strings.ToUpper(filepath.Ext(info.Name()))
	if _, exists := walkfs.inext[ext]; exists {
		return true
	} else {
		return false
	}
}

// shouldExcludePath returns true if the given relative path should be excluded
func (walkfs *WalkFS) shouldExcludePath(relpath string) bool {
	// Exclude any paths which have a .<folder> as part of their path
	if relpath != "." {
		for _, path := range strings.Split(relpath, pathSeparator) {
			if strings.HasPrefix(path, ".") {
				return true
			}
		}
	}
	// Include all files if no inclusions are specified
	if len(walkfs.expath) == 0 {
		return false
	}
	// Exclude by path prefix
	relpath = pathSeparator + strings.Trim(relpath, pathSeparator) + pathSeparator
	for path := range walkfs.expath {
		if strings.HasPrefix(relpath, path) {
			return true
		}
	}
	// Exclude by extension
	if len(walkfs.exext) == 0 {
		return false
	}
	ext := strings.ToUpper(filepath.Ext(relpath))
	if _, exists := walkfs.exext[ext]; exists {
		return true
	} else {
		return false
	}
}

// shouldExcludeFile returns true if the given file should not be visited
// based on file extension
func (walkfs *WalkFS) shouldExcludeFile(info fs.FileInfo) bool {
	// Ignore anything which isn't a regular file
	if !info.Mode().IsRegular() {
		return false
	}
	// Include all files if no inclusions are specified
	if len(walkfs.exext) > 0 {
		ext := strings.ToUpper(filepath.Ext(info.Name()))
		if _, exists := walkfs.exext[ext]; exists {
			return true
		}
	}
	if len(walkfs.exname) > 0 {
		if _, exists := walkfs.exname[info.Name()]; exists {
			return true
		}
	}
	// Return false - no exclusions
	return false
}
