package phpdist

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Version string `yaml:"version"`
}

// BuildpackYMLParser parses the buildpack.yml file for Cpython-related
// configurations.
type BuildpackYMLParser struct{}

// NewBuildpackYMLParser creates a BuildpackYMLParser
func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

// Parse decodes a given buildpack.yml file if it contains a php entry, and
// returns a struct representing the user-specified configuration.
func (p BuildpackYMLParser) Parse(path string) (Config, error) {
	var buildpack struct {
		PHP Config `yaml:"php"`
	}

	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}
	defer file.Close()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return Config{}, err
		}
	}

	return buildpack.PHP, nil
}

// ParseVersion decodes a given buildpack.yml file if it contains a php entry,
// and returns the corresponding version string.
func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	config, err := p.Parse(path)
	if err != nil {
		return "", err
	}

	return config.Version, nil
}
