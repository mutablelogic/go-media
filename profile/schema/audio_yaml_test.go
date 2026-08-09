package schema_test

import (
	"os"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	require "github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
)

const (
	yamlAudioProfile = "../../etc/profile/audio.yaml"
)

func TestNewAudioProfile_YAML(t *testing.T) {
	require := require.New(t)

	// Read the data
	data, err := os.ReadFile(yamlAudioProfile)
	require.NoError(err)

	// Parse the data
	var profiles struct {
		Audio []schema.AudioProfile `yaml:"audio"`
	}
	err = yaml.Unmarshal(data, &profiles)
	require.NoError(err)
	require.NotEmpty(profiles.Audio)
}
