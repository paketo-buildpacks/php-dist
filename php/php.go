package php

import (
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
	phpLayer         layers.DependencyLayer
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
		phpLayer:  context.Layers.DependencyLayer(dep),
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

		return c.launchLayer.WriteMetadata(layers.Metadata{
			Processes: []layers.Process{{"web", "generic php start command"}},
		})
	}, c.flags()...)
}

func (n Contributor) flags() []layers.Flag {
	var flags []layers.Flag

	if n.launchContribution {
		flags = append(flags, layers.Launch)
	}

	if n.buildContribution {
		flags = append(flags, layers.Build)
	}

	//TODO: handle cache flag if that needs to be true

	return flags
}
