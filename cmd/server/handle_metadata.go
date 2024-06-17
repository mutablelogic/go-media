package main

import (
	"encoding/json"
	"net/http"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

// POST /metadata
// Returns media file metadata
func handle_metadata(w http.ResponseWriter, r *http.Request) {
	// Always close the body
	defer r.Body.Close()

	// Check method
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read input stream
	reader, err := ffmpeg.NewReader(r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Get the metadata
	metadata := make(map[string]string)
	for _, tag := range ff.AVUtil_dict_entries(reader.Metadata()) {
		metadata[tag.Key()] = tag.Value()
	}

	// Write the metadata to the response
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metadata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
