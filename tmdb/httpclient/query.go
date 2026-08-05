package tmdb

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Query is a search query extracted from a media file's path: the title to
// search for, its release year if present, and (for a TV episode) its
// season/episode number.
type Query struct {
	Name    string
	Year    int // 0 if no year was found
	Season  int // 0 unless Episode is also set
	Episode int // 0 unless this looks like a TV episode
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	episodeRe     = regexp.MustCompile(`(?i)[\s._-]s(\d{1,2})e(\d{1,3})\b`)
	yearRe        = regexp.MustCompile(`[\s._(\[](19\d{2}|20\d{2})[\s._)\]-]?`)
	noiseRe       = regexp.MustCompile(`(?i)[\s._-](2160p|1080p|720p|480p|blu-?ray|bdrip|brrip|webrip|web-?dl|hdtv|dvdrip|remux|hdr10?|dolby.?vision|atmos|x264|x265|h\.?264|h\.?265|hevc|avc|xvid|divx|aac\d?|ac3|dts(-hd)?|truehd|\d\.\d(ch)?|extended|unrated|director.?s.?cut|theatrical|proper|repack|limited|internal|multi|dual.?audio)\b.*$`)
	sepRe         = regexp.MustCompile(`[._]+`)
	spaceRe       = regexp.MustCompile(`\s{2,}`)
	genericNameRe = regexp.MustCompile(`(?i)^(video[\s._-]?ts(?:[\s._-]?\d+)*|vts[\s._-]?\d+(?:[\s._-]?\d+)*|title[\s._-]?\d*|track[\s._-]?\d*|movie|main|cd[\s._-]?\d+|disc[\s._-]?\d+|\d+)$`)
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ExtractQuery derives a search query - title, year, and (for a TV episode)
// season/episode - from a media file's path.
func ExtractQuery(path string) Query {
	q := parseName(filepath.Base(path))
	dir := filepath.Dir(path)
	if dir == "." || dir == string(filepath.Separator) {
		return q
	}

	// Check for parent directory with a "Title (Year)" convention
	dq := parseName(filepath.Base(dir))
	if isGenericName(q.Name) && !isGenericName(dq.Name) {
		// Keep the directory's year unless the filename had its own.
		if q.Year != 0 {
			dq.Year = q.Year
		}
		q = dq
	} else if q.Year == 0 {
		q.Year = dq.Year
	}

	return q
}

// isGenericName reports whether name carries no real title information -
// either empty, or matched by genericNameRe.
func isGenericName(name string) bool {
	return name == "" || genericNameRe.MatchString(name)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// parseName extracts a Query from a single filename or directory name (not
// a full path). It doesn't attempt to strip a trailing "-GROUP"
// release-group tag on its own, since that's indistinguishable from a
// genuinely hyphenated title (e.g. "Spider-Man"); a release group is only
// removed when it trails a recognized quality/source/codec tag matched by
// noiseRe.
func parseName(name string) Query {
	var q Query

	// Remove file extension first, so we don't mistake it for a release group
	name = strings.TrimSuffix(name, filepath.Ext(name))

	// Check for a season/episode marker first
	if m := episodeRe.FindStringSubmatchIndex(name); m != nil {
		q.Season, _ = strconv.Atoi(name[m[2]:m[3]])
		q.Episode, _ = strconv.Atoi(name[m[4]:m[5]])
		name = name[:m[0]]
	}

	// Check for year
	if m := yearRe.FindStringSubmatchIndex(name); m != nil {
		q.Year, _ = strconv.Atoi(name[m[2]:m[3]])
		name = name[:m[0]]
	}

	// Cleanup
	name = noiseRe.ReplaceAllString(name, "")
	name = sepRe.ReplaceAllString(name, " ")
	name = spaceRe.ReplaceAllString(name, " ")
	q.Name = strings.TrimSpace(name)

	// Return the query even if the title is empty
	return q
}
