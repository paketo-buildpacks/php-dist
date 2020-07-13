package phpdist

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

type Environment struct {
	logger LogEmitter
}

func NewEnvironment(logger LogEmitter) Environment {
	return Environment{
		logger: logger,
	}
}

func (e Environment) Configure(layer packit.Layer) error {
	layer.SharedEnv.Prepend("PATH", filepath.Join(layer.Path, "sbin"), ":")
	layer.SharedEnv.Override("MIBDIRS", filepath.Join(layer.Path, "mibs"))
	layer.SharedEnv.Override("PHP_HOME", layer.Path)

	extensionsDir, apiVersion, err := parseExtensions(layer.Path)
	if err != nil {
		return err
	}

	layer.SharedEnv.Override("PHP_EXTENSION_DIR", extensionsDir)
	layer.SharedEnv.Override("PHP_API", apiVersion)

	e.logger.Environment(layer.SharedEnv)

	return nil
}

func parseExtensions(root string) (string, string, error) {
	folders, err := filepath.Glob(filepath.Join(root, "lib/php/extensions/no-debug-non-zts*"))
	if err != nil {
		return "", "", err
	}

	if len(folders) == 0 {
		return "", "", errors.New("php extensions dir not found")
	}

	extDir := folders[0]
	extDirChunks := strings.Split(extDir, "-")
	apiVersion := extDirChunks[len(extDirChunks)-1]

	return extDir, apiVersion, nil
}
