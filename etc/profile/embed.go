package profile

import (
	_ "embed"
)

//go:embed audio.yaml
var AudioProfilesYAML []byte

//go:embed video.yaml
var VideoProfilesYAML []byte

//go:embed format.yaml
var FormatProfilesYAML []byte
