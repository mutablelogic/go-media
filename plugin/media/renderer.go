package main

import (
	"context"
	"os"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *plugin) Mimetypes() []string {
	result := make([]string, 0, len(p.mimetypes))
	for k := range p.mimetypes {
		result = append(result, k)
	}
	return result
}

func (p *plugin) Read(ctx context.Context, path string) (Document, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, ErrBadParameter.With(path)
	}
	media, err := p.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer p.Release(media)

	// Return success
	return NewDocument(path, info, media)
}
