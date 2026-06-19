package main

import (
	"encoding/json"

	"github.com/mutablelogic/go-media/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func VersionJSON() string {
	metadata := make(map[string]string)
	for _, v := range version.Map() {
		metadata[v.Key] = v.Value
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}
