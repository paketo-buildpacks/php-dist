package phpdist

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
)

// Environment sets build- and launch-time environment variables to
// properly install and configure PHP.
type Environment struct {
}

func NewEnvironment() Environment {
	return Environment{}
}

// Configure sets build- and launch-time environment variables on the layer
// passed as input. Proper values for the environment variables are determined
// by the layer path, and the other paths provided as inputs. See PHP documentation
// for an explanation of the environment variables' significance.
func (e Environment) Configure(layer packit.Layer, extensionsDir string, defaultIniPath string, iniScanDirs []string) error {
	layer.SharedEnv.Prepend("PATH", filepath.Join(layer.Path, "bin"), string(os.PathSeparator))
	layer.SharedEnv.Prepend("PATH", filepath.Join(layer.Path, "sbin"), string(os.PathListSeparator))
	layer.SharedEnv.Default("MIBDIRS", filepath.Join(layer.Path, "mibs"))

	layer.SharedEnv.Default("PHP_HOME", layer.Path)
	layer.SharedEnv.Default("PHPRC", filepath.Dir(defaultIniPath))

	if scanDir, ok := os.LookupEnv("PHP_INI_SCAN_DIR"); ok {
		layer.SharedEnv.Default("PHP_INI_SCAN_DIR", scanDir)
	} else {
		layer.SharedEnv.Append("PHP_INI_SCAN_DIR",
			strings.Join(iniScanDirs, string(os.PathListSeparator)),
			string(os.PathListSeparator),
		)
	}

	extDirChunks := strings.Split(extensionsDir, "-")
	apiVersion := extDirChunks[len(extDirChunks)-1]

	layer.SharedEnv.Default("PHP_EXTENSION_DIR", extensionsDir)
	layer.SharedEnv.Default("PHP_API", apiVersion)

	return nil
}
