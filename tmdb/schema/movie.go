package schema

///////////////////////////////////////////////////////////////////////////////
// TYPES

// MovieDetails is a movie, with the details which a search doesn't return,
// such as its runtime
type MovieDetails struct {
	Movie

	// The runtime in minutes, which is zero when it isn't known
	Runtime uint64 `json:"runtime"`

	// The movie's IMDb identifier, such as tt0078748, when it's known
	IMDbID string `json:"imdb_id,omitempty"`
}
