package video

import (
	"context"
	"io"
	"regexp"
	"strconv"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	tmdbclient "github.com/mutablelogic/go-media/tmdb/httpclient"
	tmdbschema "github.com/mutablelogic/go-media/tmdb/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for video files: looks up the movie named by the
	// input's filename (see gomedia.NamedReader) on TMDB. Requires a TMDB
	// client (see WithTMDB); does nothing without one.
	metadata.AddHandler(regexp.MustCompile(`^video/.*$`), "tmdb", func(ctx context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		client := o.TMDB()
		if client == nil {
			return nil, nil
		}

		named, ok := r.(gomedia.NamedReader)
		if !ok || named.Name() == "" {
			return nil, nil
		}

		q := tmdbclient.ExtractQuery(named.Name())
		if q.Name == "" || q.Episode != 0 {
			// No usable title, or this looks like a TV episode: TMDB
			// search only covers movies for now.
			return nil, nil
		}

		req := tmdbschema.MovieSearchRequest{Term: q.Name}
		if q.Year > 0 {
			req.PrimaryReleaseYear = types.Ptr(strconv.Itoa(q.Year))
		}

		resp, err := client.SearchMovies(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Results) == 0 {
			return nil, nil
		}

		// TMDB's search endpoint returns results ranked by relevance; take
		// the top match.
		movie := resp.Results[0]
		entries := map[string]gomedia.Metadata{
			"tmdb:id":    meta{key: "tmdb:id", value: strconv.FormatUint(movie.ID, 10)},
			"tmdb:title": meta{key: "tmdb:title", value: movie.Title},
			"dc:title":   meta{key: "dc:title", value: movie.Title},
		}
		if movie.ReleaseDate != "" {
			entries["tmdb:releasedate"] = meta{key: "tmdb:releasedate", value: movie.ReleaseDate}
			entries["dc:date"] = meta{key: "dc:date", value: movie.ReleaseDate}
		}
		if movie.Overview != "" {
			entries["tmdb:overview"] = meta{key: "tmdb:overview", value: movie.Overview}
			entries["dc:description"] = meta{key: "dc:description", value: movie.Overview}
		}

		return metadata.FilterMetadata(entries, o), nil
	}, "tmdb", "dc")
}
