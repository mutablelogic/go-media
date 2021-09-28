package main

import (
	"context"
	"io"
	"io/fs"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
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

func (p *plugin) Read(ctx context.Context, r io.Reader, info fs.FileInfo) (Document, error) {
	media, err := p.Open(r, 0)
	if err != nil {
		return nil, err
	}
	defer p.Release(media)

	// TODO: Get path from the context
	return NewDocument("/", info, media)
}
