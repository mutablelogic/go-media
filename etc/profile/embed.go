package profile

import (
	_ "embed"
)

//go:embed profiles.yaml
var ProfilesYAML []byte
