package main

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	// Packages
	router "github.com/djthorpe/go-server/pkg/httprouter"

	// Namespace imports
	. "github.com/djthorpe/go-server"
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

type BucketFolderResponse struct {
	Name    string    `json:"name"`
	ModTime time.Time `json:"modtime"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRoutePing   = regexp.MustCompile(`^/?$`)
	reRouteBucket = regexp.MustCompile(`^/(\w+)/?$`)
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
	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, file := range files {
		wg.Add(1)
		go func(file *BucketEntry) {
			defer wg.Done()
			// Exclude files based on file extension
			if !p.handlesFile(file.Path) {
				return
			}
			// Read documents
			document, err := p.Read(req.Context(), filepath.Join(bucket.Path, file.Path))
			if err != nil {
				p.errs <- err
			} else {
				lock.Lock()
				response.Media = append(response.Media, document)
				lock.Unlock()
			}
		}(file)
	}

	// Wait for all media to be collected
	wg.Wait()

	if _, err := json.Marshal(response); err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}
