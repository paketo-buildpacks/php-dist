package phpdist

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface EnvironmentConfiguration --output fakes/environment_configuration.go
type EnvironmentConfiguration interface {
	Configure(layer packit.Layer) error
}

//go:generate faux --interface BuildPlanRefinery --output fakes/build_plan_refinery.go
type BuildPlanRefinery interface {
	BillOfMaterials(dependency postal.Dependency) packit.BuildpackPlan
}

func Build(entries EntryResolver,
	dependencies DependencyManager,
	environment EnvironmentConfiguration,
	planRefinery BuildPlanRefinery,
	logger LogEmitter,
	clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		var err error

		logger.Title(context.BuildpackInfo)
		logger.Process("Resolving PHP version")

		entry := entries.Resolve(context.Plan.Entries)

		version, _ := entry.Metadata["version"].(string)
		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)

		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency.Version)

		phpLayer, err := context.Layers.Get("php")
		if err != nil {
			return packit.BuildResult{}, err
		}

		phpLayer.Launch = entry.Metadata["launch"] == true
		phpLayer.Build = entry.Metadata["build"] == true
		phpLayer.Cache = entry.Metadata["build"] == true

		bom := planRefinery.BillOfMaterials(postal.Dependency{
			ID:      dependency.ID,
			Name:    dependency.Name,
			SHA256:  dependency.SHA256,
			Stacks:  dependency.Stacks,
			URI:     dependency.URI,
			Version: dependency.Version,
		})

		cachedSHA, ok := phpLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", phpLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Plan:   bom,
				Layers: []packit.Layer{phpLayer},
			}, nil
		}

		logger.Process("Executing build process")

		err = phpLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		phpLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		logger.Subprocess("Installing PHP %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, phpLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		err = environment.Configure(phpLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Plan:   bom,
			Layers: []packit.Layer{phpLayer},
		}, nil
	}
}
