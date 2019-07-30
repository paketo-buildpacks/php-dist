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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/php-dist-cnb/php"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
)

type BuildpackYAML struct {
	Config Config `yaml:"php"`
}

type Config struct {
	Version string `yaml:"version"`
}

func main() {
	detectionContext, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to run detection: %s", err)
		os.Exit(101)
	}

	code, err := runDetect(detectionContext)
	if err != nil {
		detectionContext.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	exists, err := helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	buildpackYAML := BuildpackYAML{}
	version := context.BuildPlan[php.Dependency].Version
	if exists {
		err = helper.ReadBuildpackYaml(buildpackYAMLPath, &buildpackYAML)
		if err != nil {
			return detect.FailStatusCode, err
		}
		if buildpackYAML.Config.Version != "" {
			version = buildpackYAML.Config.Version
		}
	}

	return context.Pass(buildplan.BuildPlan{
		php.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{},
		},
	})

}
