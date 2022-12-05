package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	// Packages
	"github.com/mutablelogic/go-media/pkg/media"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Buckets map[string]string `yaml:"buckets"`
}

type plugin struct {
	*media.Manager
	errs      chan error
	mimetypes map[string]bool
	buckets   map[string]*Bucket
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Get configuration
	cfg := media.DefaultConfig
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	}
	cfg2 := Config{}
	if err := provider.GetConfig(ctx, &cfg2); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Set up buckets
	p.buckets = make(map[string]*Bucket)
	for k, v := range cfg2.Buckets {
		if bucket, err := NewBucket(k, v); err != nil {
			provider.Print(ctx, err)
			return nil
		} else if _, exists := p.buckets[bucket.Name]; exists {
			provider.Print(ctx, ErrDuplicateEntry.With(bucket.Name))
			return nil
		} else {
			p.buckets[bucket.Name] = bucket
		}
	}

	// Create a media manager
	p.errs = make(chan error)
	if mgr, err := media.NewManagerWithConfig(cfg, p.errs); err != nil {
		provider.Print(ctx, err)
		return nil
	} else {
		p.Manager = mgr
	}

	// Enumerate mimetypes
	p.mimetypes = make(map[string]bool)
	for _, format := range p.Formats() {
		ext := strings.Split(format.Ext(), ",")
		for _, ext := range ext {
			if ext != "" {
				ext = "." + strings.TrimPrefix(ext, ".")
				key := strings.ToLower(strings.TrimSpace(ext))
				p.mimetypes[key] = true
			}
		}
		if mimetype := format.MimeType(); mimetype != "" {
			key := strings.ToLower(strings.TrimSpace(mimetype))
			p.mimetypes[key] = true
		}
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<media"
	if len(p.buckets) > 0 {
		str += fmt.Sprint(" buckets=", p.Buckets())
	}
	str += fmt.Sprint(" ", p.Manager)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "media"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	// Add handlers
	if err := p.AddHandlers(ctx, provider); err != nil {
		return err
	}
	// Run until cancelled - print any errors from media manager
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case err := <-p.errs:
			if err != nil {
				provider.Print(ctx, err)
			}
		}
	}

	// Close the pool
	if err := p.Manager.Close(); err != nil {
		provider.Print(ctx, err)
	}

	// Close error channel
	close(p.errs)

	// Return success
	return nil
}

func (p *plugin) Buckets() []*Bucket {
	result := make([]*Bucket, 0, len(p.buckets))
	for _, b := range p.buckets {
		result = append(result, b)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *plugin) handlesFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	_, exists := p.mimetypes[ext]
	return exists
}
