package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	// Namespace imports
	. "github.com/djthorpe/go-errors"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type VisitFunc func(context.Context, string, fs.FileInfo) error

type WalkFS struct {
	ext     map[string]bool
	count   int
	visitfn VisitFunc
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewWalkFS(fn VisitFunc) *WalkFS {
	walkfs := new(WalkFS)
	walkfs.ext = make(map[string]bool)
	walkfs.visitfn = fn
	return walkfs
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (walkfs *WalkFS) Count() int {
	return walkfs.count
}

func (walkfs *WalkFS) IncludeExt(ext string) {
	ext = strings.TrimSpace(ext)
	ext = strings.ToUpper("." + strings.TrimPrefix(ext, "."))
	if ext != "." {
		walkfs.ext[ext] = true
	}
}

func (walkfs *WalkFS) WalkArgs(ctx context.Context, args ...string) error {
	var result error
	for _, path := range args {
		if abspath, err := filepath.Abs(path); err != nil {
			result = multierror.Append(result, ErrNotFound.With(path))
		} else {
			path = abspath
		}
		if stat, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			result = multierror.Append(result, ErrNotFound.With(path))
		} else if err != nil {
			result = multierror.Append(result, err)
		} else if stat.IsDir() {
			if err := walkfs.walk(ctx, path); err != nil {
				result = multierror.Append(result, err)
			}
		} else if stat.Mode().IsRegular() {
			if err := walkfs.visit(ctx, path, stat); err != nil {
				result = multierror.Append(result, err)
			}
		} else {
			result = multierror.Append(result, ErrNotFound.With(path))
		}
	}

	// Return any errors
	return result
}

func (walkfs *WalkFS) walk(ctx context.Context, path string) error {
	// Walk filesystem
	err := filepath.WalkDir(path, func(path string, file fs.DirEntry, err error) error {
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
		if info, err := file.Info(); err == nil {
			if info.Mode().IsRegular() {
				walkfs.visit(ctx, path, info)
			}
			return nil
		}
		// Return any context error
		return ctx.Err()
	})

	// Return errors unless they are cancel/timeout
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	} else {
		return err
	}
}

func (walkfs *WalkFS) visit(ctx context.Context, path string, info fs.FileInfo) error {
	walkfs.count++

	// Exclude files by extension
	if len(walkfs.ext) > 0 {
		ext := strings.ToUpper(filepath.Ext(path))
		if _, exists := walkfs.ext[ext]; !exists {
			return nil
		}
	}

	// Call function
	if walkfs.visitfn != nil {
		return walkfs.visitfn(ctx, path, info)
	} else {
		return nil
	}
}
