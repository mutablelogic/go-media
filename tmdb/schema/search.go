package schema

import (
	"net/url"
	"strconv"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type MovieSearchRequest struct {
	Term               string  `json:"query" arg:"" name:"query" help:"Search term."`
	IncludeAdult       *bool   `json:"include_adult,omitempty" negatable:"" help:"Include adult content in search results."`
	Language           *string `json:"language,omitempty" help:"Language for the search results."`
	PrimaryReleaseYear *string `json:"primary_release_year,omitempty" help:"Filter results by primary release year."`
	Page               *uint   `json:"page,omitempty" help:"Page number of the results."`
	Region             *string `json:"region,omitempty" help:"Filter results by region."`
	Year               *string `json:"year,omitempty" help:"Filter results by year."`
}

type MovieSearchResponse struct {
	Page         uint64 `json:"page"`
	Results      []Movie
	TotalPages   uint64 `json:"total_pages"`
	TotalResults uint64 `json:"total_results"`
}

type Movie struct {
	ID               uint64   `json:"id"`
	Title            string   `json:"title"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title"`
	ReleaseDate      string   `json:"release_date"`
	Overview         string   `json:"overview"`
	Genres           []uint64 `json:"genre_ids"`
	Video            bool     `json:"video,omitempty"`

	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`

	Popularity  float32 `json:"popularity,omitempty"`
	VoteAverage float32 `json:"vote_average,omitempty"`
	VoteCount   uint64  `json:"vote_count,omitempty"`

	Adult    bool `json:"adult,omitempty"`
	SoftCore bool `json:"softcore,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r *MovieSearchRequest) Query() url.Values {
	response := url.Values{}
	if r.Term != "" {
		response.Set("query", r.Term)
	}
	if r.IncludeAdult != nil {
		response.Set("include_adult", strconv.FormatBool(types.Value(r.IncludeAdult)))
	}
	if language := types.Value(r.Language); language != "" {
		response.Set("language", language)
	}
	if year := types.Value(r.PrimaryReleaseYear); year != "" {
		response.Set("primary_release_year", year)
	}
	if r.Page != nil {
		response.Set("page", strconv.FormatUint(uint64(types.Value(r.Page)), 10))
	}
	if region := types.Value(r.Region); region != "" {
		response.Set("region", region)
	}
	if year := types.Value(r.Year); year != "" {
		response.Set("year", year)
	}
	return response
}
