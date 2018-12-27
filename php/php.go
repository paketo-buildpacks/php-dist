package php

import (
	"fmt"
	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"path/filepath"
)

const Dependency = "php"

type Contributor struct {
	app                application.Application
	launchContribution bool
	launchLayer        layers.Layers
	httpdLayer         layers.DependencyLayer
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
		httpdLayer:  context.Layers.DependencyLayer(dep),
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	return c.httpdLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		if err := layer.OverrideLaunchEnv("APP_ROOT", c.app.Root); err != nil {
			return err
		}

		if err := layer.OverrideLaunchEnv("SERVER_ROOT", layer.Root); err != nil {
			return err
		}

		return c.launchLayer.WriteMetadata(layers.Metadata{
			Processes: []layers.Process{{"web", fmt.Sprintf("httpd -f %s -k start -DFOREGROUND", filepath.Join(c.app.Root,"httpd.conf"))}},
		})
	}, c.flags()...)
}

func (n Contributor) flags() []layers.Flag {
	var flags []layers.Flag

	if n.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}
