/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package php

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	yaml "gopkg.in/yaml.v2"
)

// Dependency is a build plan dependency indicating a requirement for PHP binaries
const Dependency = "php-binary"

// Version returns the selected version of PHP using the following precedence:
//
// 1. `php.verison` from `buildpack.yml`
// 2. Build Plan Version
// 3. Buildpack Metadata "default_version"
// 4. Empty string
func Version(buildpackYAML BuildpackYAML, buildpack buildpack.Buildpack, dependency buildplan.Dependency) string {
	if buildpackYAML.Config.Version != "" {
		return buildpackYAML.Config.Version
	}

	if dependency.Version != "" {
		return dependency.Version
	}

	if version, ok := buildpack.Metadata["default_version"].(string); ok {
		return version
	}

	return ""
}

// API returns the API string for the given PHP version
func API(version string) string {
	if strings.HasPrefix(version, "7.0") {
		return "20151012"
	} else if strings.HasPrefix(version, "7.1") {
		return "20160303"
	} else if strings.HasPrefix(version, "7.2") {
		return "20170718"
	} else if strings.HasPrefix(version, "7.3") {
		return "20180731"
	} else {
		return ""
	}
}

// BuildpackYAML represents user specified config options through `buildpack.yml`
type BuildpackYAML struct {
	Config Config `yaml:"php"`
}

// Config represents PHP specific configuration options for BuildpackYAML
type Config struct {
	Version      string `yaml:"version"`
	WebServer    string `yaml:"webserver"`
	WebDirectory string `yaml:"webdirectory"`
	Script       string `yaml:"script"`
}

// LoadBuildpackYAML reads `buildpack.yml` and PHP specific config options in it
func LoadBuildpackYAML(appRoot string) (BuildpackYAML, error) {
	buildpackYAML, configFile := BuildpackYAML{}, filepath.Join(appRoot, "buildpack.yml")
	if exists, err := helper.FileExists(configFile); err != nil {
		return BuildpackYAML{}, err
	} else if exists {
		file, err := os.Open(configFile)
		if err != nil {
			return BuildpackYAML{}, err
		}
		defer file.Close()

		contents, err := ioutil.ReadAll(file)
		if err != nil {
			return BuildpackYAML{}, err
		}

		err = yaml.Unmarshal(contents, &buildpackYAML)
		if err != nil {
			return BuildpackYAML{}, err
		}
	}
	return buildpackYAML, nil
}
