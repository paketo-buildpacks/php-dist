package phpdist

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

// PHPIniConfig represents the data that will be inserted in a templated
// php.ini file.
type PhpIniConfig struct {
	IncludePath string
	// include_path = "{{.PhpHome}}/lib/php:{{.AppRoot}}/{{.LibDirectory}}"
	ExtensionDir string
	// extension_dir = "{{.PhpHome}}/lib/php/extensions/no-debug-non-zts-{{.PhpAPI}}"
	Extensions     []string
	ZendExtensions []string
}

// PHPFileManager finds, copies, and creates files necessary for a proper PHP
// installation.
type PHPFileManager struct{}

func NewPHPFileManager() PHPFileManager {
	return PHPFileManager{}
}

// FindExtensions checks a path relative to the layer passed as input
// where it expects extensions to be pre-installed. It fails if a directory
// in the expected location does not exist.
func (f PHPFileManager) FindExtensions(layerRoot string) (string, error) {
	folders, err := filepath.Glob(filepath.Join(layerRoot, "lib/php/extensions/no-debug-non-zts-*"))
	if err != nil {
		return "", err
	}

	if len(folders) != 1 {
		return "", fmt.Errorf("expected 1 PHP extensions dir matching '%s', but found %d", filepath.Join(layerRoot, "lib/php/extensions/no-debug-non-zts-*"), len(folders))
	}

	return folders[0], nil
}

// WriteConfig generates a default PHP configuration and stores the resulting *.ini
// files in the layer provided as input.
func (f PHPFileManager) WriteConfig(layerRoot, cnbPath string, data PhpIniConfig) (string, string, error) {
	err := os.MkdirAll(filepath.Join(layerRoot, "etc"), os.ModePerm)
	if err != nil {
		return "", "", err
	}

	defaultConfig := filepath.Join(layerRoot, "etc", "php.ini")
	err = fs.Copy(filepath.Join(cnbPath, "config", "default.ini"), defaultConfig)
	if err != nil {
		return "", "", err
	}

	tmpl, err := template.New("buildpack.ini").ParseFiles(filepath.Join(cnbPath, "config", "buildpack.ini"))
	if err != nil {
		return "", "", fmt.Errorf("failed to parse php.ini template: %w", err)
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		// not tested
		return "", "", err
	}

	buildpackConfig, err := os.OpenFile(filepath.Join(layerRoot, "etc", "buildpack.ini"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", "", err
	}
	defer func() {
		if err := buildpackConfig.Close(); err != nil {
			// Log the error or handle it appropriately
			fmt.Fprintf(os.Stderr, "failed to close buildpackConfig: %v\n", err)
		}
	}()

	_, err = io.Copy(buildpackConfig, &b)
	if err != nil {
		// not tested
		return "", "", err
	}

	return defaultConfig, buildpackConfig.Name(), nil
}
