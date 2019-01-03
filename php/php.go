package php

import (
	"fmt"
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const Dependency = "php"

type Contributor struct {
	app                application.Application
	launchContribution bool
	buildContribution  bool
	launchLayer        layers.Layers
	phpLayer           layers.DependencyLayer
}

func NewContributor(context build.Build) (c Contributor, willContribute bool, err error) {
	plan, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	deps, err := context.Buildpack.Dependencies()
	if err != nil {
		return Contributor{}, false, err
	}

	dep, err := deps.Best(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{
		app:         context.Application,
		launchLayer: context.Layers,
		phpLayer:    context.Layers.DependencyLayer(dep),
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	return c.phpLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("PHPRC", filepath.Join(layer.Root, "etc")); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("MIBDIRS", filepath.Join(layer.Root, "mibs")); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("PHP_INI_SCAN_DIR", filepath.Join(c.app.Root, "etc", "php.ini.d")); err != nil {
			return err
		}

		// TODO: How do we know when to use php-fpm or not?
		isWebApp, err := helper.FileExists(filepath.Join(c.app.Root, "htdocs"))
		if err != nil {
			return err
		}

		var procs layers.Processes

		if isWebApp {
			procs = append(procs, layers.Process{"web", fmt.Sprintf("php -S 0.0.0.0:8080 -t %s/htdocs", c.app.Root)})
		} else {
			hasMain, err := helper.FileExists(filepath.Join(c.app.Root, "main.php"))
			if err != nil {
				return err
			}

			if !hasMain {
				layer.Logger.Info("WARNING: main.php start script not found. App will not start unless you specify a custom start command.")
			} else {
				procs = append(procs, layers.Process{"web", "php main.php"})
			}
		}

		return c.launchLayer.WriteMetadata(layers.Metadata{
			Processes: procs,
		})
	}, c.flags()...)
}

func (n Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if n.launchContribution {
		flags = append(flags, layers.Launch)
	}

	if n.buildContribution {
		flags = append(flags, layers.Build)
	}

	return flags
}
