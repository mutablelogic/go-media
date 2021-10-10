package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	media "github.com/mutablelogic/go-media/pkg/media"
	router "github.com/mutablelogic/go-server/pkg/httprouter"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PingResponse struct {
	Buckets []*Bucket `json:"buckets"`
}

type BucketRequest struct {
	Path string `json:"path"`
}

type BucketResponse struct {
	Bucket  string         `json:"bucket"`
	Path    string         `json:"path,omitempty"`
	Name    string         `json:"name,omitempty"`
	Folders []*BucketEntry `json:"folders,omitempty"`
	Media   []Document     `json:"media,omitempty"`
}

type ArtworkResponse struct {
	Bucket string         `json:"bucket"`
	Path   string         `json:"path,omitempty"`
	Name   string         `json:"name,omitempty"`
	Media  []ArtworkMedia `json:"media,omitempty"`
	hash   map[string]string
}

type ArtworkMedia struct {
	Mimetype string `json:"mimetype"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
	Index    int    `json:"index"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRoutePing    = regexp.MustCompile(`^/?$`)
	reRouteBucket  = regexp.MustCompile(`^/(\w+)/?$`)
	reRouteArtwork = regexp.MustCompile(`^/(\w+)/artwork/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (p *plugin) AddHandlers(ctx context.Context, provider Provider) error {
	// Add handler for ping
	if err := provider.AddHandlerFuncEx(ctx, reRoutePing, p.ServePing); err != nil {
		return err
	}
	// Add handler for bucket
	if err := provider.AddHandlerFuncEx(ctx, reRouteBucket, p.ServeBucket, http.MethodGet, http.MethodPost); err != nil {
		return err
	}
	// Add handler for artwork
	if err := provider.AddHandlerFuncEx(ctx, reRouteArtwork, p.ServeArtwork, http.MethodGet, http.MethodPost); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (p *plugin) ServePing(w http.ResponseWriter, req *http.Request) {
	// Populate response
	response := PingResponse{
		Buckets: p.Buckets(),
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

func (p *plugin) ServeArtwork(w http.ResponseWriter, req *http.Request) {
	// Decode params, params[0] is the bucket name
	params := router.RequestParams(req)
	bucket, exists := p.buckets[params[0]]
	if !exists {
		router.ServeError(w, http.StatusNotFound)
		return
	}

	// Obtain query parameters
	var query BucketRequest
	if req.Method == http.MethodPost {
		if err := router.RequestBody(req, &query); err != nil {
			router.ServeError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else if req.Method == http.MethodGet {
		if err := router.RequestQuery(req, &query); err != nil {
			router.ServeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Populate response
	response := ArtworkResponse{
		Bucket: bucket.Name,
		Path:   bucket.Path,
		hash:   make(map[string]string),
	}

	// Set name
	if query.Path != "" {
		response.Name = strings.TrimPrefix(filepath.Base(query.Path), PathSeparator)
	}

	// Get files
	files, err := bucket.FilesForPath(query.Path)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Extract files based on mimetypes
	var lock sync.Mutex
	if err := p.process(bucket, files, func(path string) error {
		// Read media
		media, err := p.OpenFile(path)
		if err != nil {
			return err
		}
		defer p.Release(media)
		if !media.Flags().Is(MEDIA_FLAG_ARTWORK) {
			return nil
		}
		// Relative path to media
		relpath, err := filepath.Rel(bucket.Path, path)
		if err != nil {
			return err
		}

		// Cycle through streams to obtain artwork
		for _, stream := range media.Streams() {
			if stream.Flags().Is(MEDIA_FLAG_ARTWORK) {
				lock.Lock()
				if media := response.process(relpath, media, stream); media != nil {
					response.Media = append(response.Media, *media)
				}
				lock.Unlock()
			}
		}
		// Return success
		return nil
	}); err != nil {
		p.errs <- err
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

func (p *plugin) ServeBucket(w http.ResponseWriter, req *http.Request) {
	// Decode params, params[0] is the bucket name
	params := router.RequestParams(req)
	bucket, exists := p.buckets[params[0]]
	if !exists {
		router.ServeError(w, http.StatusNotFound)
		return
	}

	// Obtain query parameters
	var query BucketRequest
	if req.Method == http.MethodPost {
		if err := router.RequestBody(req, &query); err != nil {
			router.ServeError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else if req.Method == http.MethodGet {
		if err := router.RequestQuery(req, &query); err != nil {
			router.ServeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Populate response
	response := BucketResponse{
		Bucket: bucket.Name,
		Path:   bucket.Path,
	}

	// Get folders
	if folders, err := bucket.FoldersForPath(query.Path); err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		response.Folders = folders
	}

	// Set name
	if query.Path != "" {
		response.Name = strings.TrimPrefix(filepath.Base(query.Path), PathSeparator)
	}

	// Get files
	files, err := bucket.FilesForPath(query.Path)
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Extract files based on mimetypes
	var lock sync.Mutex
	if err := p.process(bucket, files, func(path string) error {
		// Get info
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		// Open file
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()

		// Read documents
		document, err := p.Read(req.Context(), r, info, nil)
		if err != nil {
			return err
		}

		// Append document to response
		lock.Lock()
		response.Media = append(response.Media, document)
		lock.Unlock()

		// Return success
		return nil
	}); err != nil {
		p.errs <- err
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *plugin) process(bucket *Bucket, files []*BucketEntry, fn func(abspath string) error) error {
	var result error
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file *BucketEntry) {
			defer wg.Done()
			// Exclude files based on file extension
			if !p.handlesFile(file.Path) {
				return
			}
			// Call function
			if err := fn(filepath.Join(bucket.Path, file.Path)); err != nil {
				result = multierror.Append(result, err)
			}
		}(file)
	}

	// Wait for all media to be collected
	wg.Wait()

	// Return any errors
	return result
}

func (r *ArtworkResponse) process(path string, media *media.MediaInput, stream *media.Stream) *ArtworkMedia {
	// Get artwork and hashcode from the stream
	bytes := stream.Artwork()
	if bytes == nil {
		return nil
	}

	// Get hashcode for the artwork, only process if not already processed
	key := fmt.Sprintf("%x", md5.Sum(bytes))
	if _, exists := r.hash[key]; exists {
		return nil
	} else {
		r.hash[key] = key
	}

	// Append artwork
	return &ArtworkMedia{
		Mimetype: http.DetectContentType(bytes),
		Size:     int64(len(bytes)),
		Path:     PathSeparator + strings.TrimPrefix(path, PathSeparator),
		Index:    stream.Index(),
	}
}
