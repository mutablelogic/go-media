package cmd

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	Recursive  bool
	ExcludeExt map[string]struct{}
	Template   *Templater
}

type WalkOpt func(*opt) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func apply(opts []WalkOpt) (*opt, error) {
	o := new(opt)
	o.ExcludeExt = make(map[string]struct{})

	// Apply options
	for _, fn := range opts {
		if err := fn(o); err != nil {
			return nil, err
		}
	}

	// Return success
	return o, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OPTIONS

func WithRecursive() WalkOpt {
	return func(o *opt) error {
		o.Recursive = true
		return nil
	}
}

func WithExcludeExt(exts ...string) WalkOpt {
	return func(o *opt) error {
		for i, ext := range exts {
			if ext = strings.TrimSpace(ext); !strings.HasPrefix(ext, ".") {
				return gomedia.ErrBadParameter.With("extension must start with a dot: " + ext)
			} else {
				exts[i] = strings.ToLower(ext)
			}
		}
		for _, ext := range exts {
			o.ExcludeExt[ext] = struct{}{}
		}
		return nil
	}
}

func WithTemplate(t string) WalkOpt {
	return func(o *opt) error {
		if tmpl, err := NewTemplater(t); err != nil {
			return err
		} else {
			o.Template = tmpl
		}
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WalkFS(ctx context.Context, root fs.FS, fn func(context.Context, string, fs.DirEntry, *Templater) error, opts ...WalkOpt) error {
	// Gather options
	o, err := apply(opts)
	if err != nil {
		return err
	}

	// Walk the filesystem
	return fs.WalkDir(root, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip if non-recursive and not the root directory, or if the file is hidden (starts with a dot)
		if info.IsDir() {
			if path == "." {
				return nil
			}
			if !o.Recursive {
				return fs.SkipDir
			}
			return fn(ctx, path, info, nil)
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Skip excluded extensions
		if ext := strings.ToLower(filepath.Ext(info.Name())); ext != "" {
			if _, ok := o.ExcludeExt[ext]; ok {
				return nil
			}
		}

		// Callback with the path and info
		return fn(ctx, path, info, o.Template)
	})
}
