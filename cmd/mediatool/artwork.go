package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"

	// Packages
	media "github.com/mutablelogic/go-media/pkg/media"
	multierror "github.com/hashicorp/go-multierror"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GetArtwork = Command{
		Keyword:     "artwork",
		Syntax:      "<file>...",
		Description: "Extract artwork for one or more files",
		Fn:          GetArtworkFn,
	}
)

type Artwork struct {
	*media.Manager
	*WalkFS
	hash      map[string]bool
	outpath   string
	overwrite bool
	n         int
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewArtwork(cfg media.Config, errs chan<- error, outpath string, overwrite bool) (*Artwork, error) {
	artwork := new(Artwork)
	mgr, err := media.NewManagerWithConfig(cfg, errs)
	if err != nil {
		return nil, err
	}
	artwork.Manager = mgr
	artwork.WalkFS = NewWalkFS(artwork.Visit)
	artwork.hash = make(map[string]bool)
	artwork.outpath = outpath
	artwork.overwrite = overwrite
	return artwork, nil
}

func (a *Artwork) Close() error {
	var result error
	if err := a.Manager.Close(); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// COMMAND

func GetArtworkFn(ctx context.Context, cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrBadParameter.With("Missing filename argument")
	}

	// Path is current directory
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Make artwork, file walker
	artwork, err := NewArtwork(media.DefaultConfig, cmd.Errs, pwd, true)
	if err != nil {
		return err
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
				fmt.Printf("   ...%d files parsed, %d artwork written\n", artwork.Count(), artwork.n)
			}
		}
	}()

	// Walk filesystem, appending files
	result := artwork.WalkArgs(ctx, args...)

	// End reporting
	cancel()
	wg.Wait()

	// Return any errors
	return result
}

func (a *Artwork) Visit(ctx context.Context, path string, info fs.FileInfo) error {
	f, err := a.OpenFile(path)
	if err != nil {
		return err
	}
	defer a.Manager.Release(f)

	// Check for artwork flag
	if !f.Flags().Is(MEDIA_FLAG_ARTWORK) {
		return nil
	}

	// Enumerate streams to get artwork
	var result error
	for _, stream := range f.Streams() {
		if stream.Flags().Is(MEDIA_FLAG_ARTWORK) {
			if err := a.Process(ctx, f, stream, info); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return nil
}

func (a *Artwork) Process(ctx context.Context, file *media.MediaInput, stream *media.Stream, info fs.FileInfo) error {
	// Get artwork and hashcode from the stream
	bytes := stream.Artwork()
	if bytes == nil {
		return nil
	}

	// Get hashcode for the artwork, only process if not already processed
	key := fmt.Sprintf("%x", md5.Sum(bytes))
	if _, exists := a.hash[key]; exists {
		return nil
	}

	// Determine mimetype for the artwork, then the file extension
	mimetype := http.DetectContentType(bytes)
	if !strings.HasPrefix(mimetype, "image/") {
		return nil
	}
	exts, err := mime.ExtensionsByType(mimetype)
	if err != nil || len(exts) == 0 {
		return nil
	} else {
		exts = exts[len(exts)-1:]
	}

	// Determine filename
	filename := fmt.Sprintf("%s%s", strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())), exts[0])
	if file.Flags().Is(MEDIA_FLAG_ALBUM) {
		album := file.Metadata().Value(MEDIA_KEY_ALBUM)
		artist := file.Metadata().Value(MEDIA_KEY_ALBUM_ARTIST)
		if artist != nil && album != nil {
			filename = fmt.Sprintf("%s - %s%s", artist, album, exts[0])
		} else if album != nil {
			filename = fmt.Sprintf("%s%s", album, exts[0])
		}
	} else if file.Flags().Is(MEDIA_FLAG_TVSHOW) {
		tvshow := file.Metadata().Value(MEDIA_KEY_SHOW)
		if tvshow != nil {
			filename = fmt.Sprintf("%s%s", tvshow, exts[0])
		}
	} else if title := file.Metadata().Value(MEDIA_KEY_TITLE); title != nil {
		filename = fmt.Sprintf("%s%s", title, exts[0])
	}

	// Write artwork to disk
	var result error
	if err := a.Write(ctx, filename, bytes); err != nil {
		result = multierror.Append(result, err)
	} else {
		a.n++
	}
	a.hash[key] = true

	return result
}

func (a *Artwork) Write(ctx context.Context, filename string, data []byte) error {
	var path string
	var n int
	newfilename := filename
	for {
		path = filepath.Join(a.outpath, newfilename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		} else if err != nil {
			return err
		} else if a.overwrite {
			break
		}
		// Re-adjust the filename
		ext := filepath.Ext(filename)
		n = n + 1
		newfilename = strings.TrimSuffix(filename, ext) + fmt.Sprint("-", n) + ext
	}
	// We write the file here
	return ioutil.WriteFile(path, data, 0644)
}
