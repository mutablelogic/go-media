package chromaprint

import (
	"encoding/json"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Response struct {
	Status  string           `json:"status"`
	Error   ResponseError    `json:"error"`
	Results []*ResponseMatch `json:"results"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseMatch struct {
	Id         string              `json:"id"`
	Score      float64             `json:"score"`
	Recordings []ResponseRecording `json:"recordings,omitempty"`
}

type ResponseRecording struct {
	Id            string           `json:"id"`
	Title         string           `json:"title,omitempty"`
	Duration      float64          `json:"duration,omitempty"`
	Artists       []ResponseArtist `json:"artists,omitempty"`
	ReleaseGroups []ResponseGroup  `json:"releasegroups,omitempty"`
}

type ResponseArtist struct {
	Id   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type ResponseGroup struct {
	Id       string            `json:"id"`
	Type     string            `json:"type,omitempty"`
	Title    string            `json:"title,omitempty"`
	Releases []ResponseRelease `json:"releases,omitempty"`
}

type ResponseRelease struct {
	Id      string           `json:"id"`
	Mediums []ResponseMedium `json:"mediums,omitempty"`
}

type ResponseMedium struct {
	Format     string          `json:"format"`
	Position   uint            `json:"position"`
	TrackCount uint            `json:"track_count"`
	Tracks     []ResponseTrack `json:"tracks,omitempty"`
}

type ResponseTrack struct {
	Id       string           `json:"id"`
	Artists  []ResponseArtist `json:"artists,omitempty"`
	Position uint             `json:"position"`
	Title    string           `json:"title,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r *ResponseMatch) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}
