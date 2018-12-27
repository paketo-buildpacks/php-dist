package main

import (
	"fmt"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/php-cnb/php"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/detect"
)

type BuildpackYAML struct {
	Config php.Config `yaml:"httpd"`
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
	//TODO: update this search to be recursive
	matchList, err := filepath.Glob(filepath.Join(context.Application.Root, "htdocs", "*.php"))
	if err != nil {
		return context.Fail(), err
	}
	exists := len(matchList) > 0

	if !exists {
		return context.Fail(), fmt.Errorf("unable to find htdocs/index.php")
	}

	buildpackYAML, configFile := BuildpackYAML{}, filepath.Join(context.Application.Root, "buildpack.yml")
	if exists, err := helper.FileExists(configFile); err != nil {
		return context.Fail(), err
	} else if exists {
		file, err := os.Open(configFile)
		if err != nil {
			return context.Fail(), err
		}
		defer file.Close()

		contents, err := ioutil.ReadAll(file)
		if err != nil {
			return context.Fail(), err
		}

		err = yaml.Unmarshal(contents, &buildpackYAML)
		if err != nil {
			return context.Fail(), err
		}
	}

	return context.Pass(buildplan.BuildPlan{
		php.Dependency: buildplan.Dependency{
			Version:  buildpackYAML.Config.Version,
			Metadata: buildplan.Metadata{"launch": true},
		},
	})
}
