package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	file "github.com/mutablelogic/go-media/pkg/file"
)

var (
	mapHash = make(map[string]string)
)

// ProcessMedia processes media files through the pipeline
func ProcessMedia(ctx context.Context, out string, media Media) error {
	var artwork [][]byte
	var result error

	// Gather artwork
	for _, stream := range media.Streams() {
		if stream.Flags().Is(MEDIA_FLAG_ARTWORK) {
			artwork = append(artwork, stream.Artwork())
		}
	}

	// Iterate through the artwork, detect the image type and write to disk
	for _, data := range artwork {
		// Detect image type, ignore if not an image
		mimetype, ext, err := file.MimeType(data)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		} else if !strings.HasPrefix(mimetype, "image/") {
			continue
		}

		// Create file, don't overwrite existing files
		if err := WriteFile(out, media, ext, data, Hash(data), 0); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return success
	return nil
}

// Writes a file to disk, but only if it doesn't already exist
func WriteFile(path string, media Media, ext string, data []byte, hash string, i int) error {
	filename := filepath.Join(path, Filename(media, i, ext))

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// The happy path - write the file
		fmt.Println("Writing: ", filename)
		if err := ioutil.WriteFile(filename, data, 0644); err != nil {
			return err
		}
		// Add the hash to the map
		mapHash[filename] = hash
		return nil
	} else if err != nil {
		// Unexpected error
		return err
	}

	// File exists on the filesystem, check if we have computed the hash
	if existing, exists := mapHash[filename]; exists && existing == hash {
		return nil
	}

	// Compute the hash
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	} else {
		// Add the hash to the map
		mapHash[filename] = Hash(data)
	}

	// Check again if the hash matches
	if mapHash[filename] == hash {
		return nil
	}

	// Try and write again with the next index
	return WriteFile(path, media, ext, data, hash, i+1)
}

// Returns an MD5 hash string for []byte
func Hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// Creates a filename for the artwork from the media item, stream index and extension
func Filename(media Media, index int, ext string) string {
	title := strings.TrimSpace(strings.Trim(Title(media), "."))
	// Replace : and / characters
	title = strings.Replace(title, ":", "-", -1)
	title = strings.Replace(title, "/", "-", -1)

	// Return filename
	if index > 0 {
		return fmt.Sprint(title, " - ", index, ext)
	} else {
		return title + ext
	}
}

func Title(media Media) string {
	metadata := media.Metadata()
	if title := TitleAlbum(metadata); title != "" {
		return title
	}
	//if media.Flags().Is(MEDIA_FLAG_ALBUM) {
	//
	//}
	base := filepath.Base(media.URL())
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}

func TitleAlbum(metadata Metadata) string {
	var parts []string
	if compilation, ok := metadata.Value(MEDIA_KEY_COMPILATION).(bool); ok && !compilation {
		if artist, ok := metadata.Value(MEDIA_KEY_ALBUM_ARTIST).(string); ok && artist != "" {
			parts = append(parts, artist)
		} else if artist, ok := metadata.Value(MEDIA_KEY_ARTIST_SORT).(string); ok && artist != "" {
			parts = append(parts, artist)
		}
	}
	if title, ok := metadata.Value(MEDIA_KEY_ALBUM).(string); ok && title != "" {
		parts = append(parts, title)
	} else if title, ok := metadata.Value(MEDIA_KEY_ALBUM_SORT).(string); ok && title != "" {
		parts = append(parts, title)
	}
	return strings.Join(parts, " - ")
}
