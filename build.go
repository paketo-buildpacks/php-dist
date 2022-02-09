package phpdist

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface FileManager --output fakes/file_manager.go

// FileManager defines the interface for manipulating files in the PHP installation
// in the build container.
type FileManager interface {
	FindExtensions(layerRoot string) (string, error)
	WriteConfig(layerRoot, cnbPath string, data PhpIniConfig) (defaultConfig string, buildpackConfig string, err error)
}

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go

// EntryResolver defines the interface for picking the most relevant entry from
// the Buildpack Plan entries.
type EntryResolver interface {
	Resolve(name string, entries []packit.BuildpackPlanEntry, priorities []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(name string, entries []packit.BuildpackPlanEntry) (launch, build bool)
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go

// DependencyManager defines the interface for picking the best matching
// dependency, installing it, and generating a BOM.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface EnvironmentConfiguration --output fakes/environment_configuration.go

// EnvironmentConfiguration defines the interface for setting build- and launch-time
// environment variables on the layer.
type EnvironmentConfiguration interface {
	Configure(layer packit.Layer, extensionsDir, defaultIni string, scanDirs []string) error
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right php dependency to install, install it in a layer,
// and generate a Bill-of-Materials. On rebuilds, it reuses the cached
// dependency if the SHA256 of the requested version matches the SHA256 of the
// cached version. Build also sets up a default php.ini configuration.
func Build(entryResolver EntryResolver,
	dependencies DependencyManager,
	files FileManager,
	environment EnvironmentConfiguration,
	logger scribe.Emitter,
	clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		var err error

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)
		logger.Process("Resolving PHP version")

		entry, entries := entryResolver.Resolve(PHPDependency, context.Plan.Entries, EntryPriorities)
		logger.Candidates(entries)

		version, _ := entry.Metadata["version"].(string)
		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)

		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logger.Subprocess("WARNING: Setting the PHP version through buildpack.yml will be deprecated in %s v%s.", context.BuildpackInfo.Name, nextMajorVersion.String())
			logger.Subprocess("In versions >= v%s, use the $BP_PHP_VERSION environment variable to specify a version.", nextMajorVersion.String())
			logger.Break()
		}

		logger.Debug.Process("Getting the layer associated with PHP:")
		phpLayer, err := context.Layers.Get(PHPDependency)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Subprocess(phpLayer.Path)
		logger.Debug.Break()

		logger.Debug.Process("Generating the SBOM")
		logger.Debug.Break()
		bom := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := entryResolver.MergeLayerTypes(PHPDependency, context.Plan.Entries)

		phpLayer.Launch, phpLayer.Build, phpLayer.Cache = launch, build, build

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = bom
		}

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = bom
		}

		cachedSHA, ok := phpLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logger.Process("Reusing cached layer %s", phpLayer.Path)
			logger.Debug.Subprocess("SHA256 of cached PHP dependency matches SHA256 of resolved dependency")
			logger.Break()

			if phpLayer.Build {
				logger.Debug.Process("PHP layer will be available to other buildpacks during build")
			}
			if phpLayer.Launch {
				logger.Debug.Process("PHP layer will be available at runtime")
			}
			if phpLayer.Cache {
				logger.Debug.Process("PHP layer will be cached")
			}

			return packit.BuildResult{
				Layers: []packit.Layer{phpLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logger.Process("Executing build process")

		phpLayer, err = phpLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		phpLayer.Launch, phpLayer.Build, phpLayer.Cache = launch, build, build

		phpLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		logger.Subprocess("Installing PHP %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			logger.Debug.Subprocess("Installation path: %s", phpLayer.Path)
			logger.Debug.Subprocess("Dependency URI: %s", dependency.URI)
			return dependencies.Deliver(dependency, context.CNBPath, phpLayer.Path, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.Debug.Subprocess("Finding PHP extensions directory")
		extensionsDir, err := files.FindExtensions(phpLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Break()

		libDir := filepath.Join("lib")
		if userLibDir := os.Getenv("BP_PHP_LIB_DIR"); userLibDir != "" {
			libDir = userLibDir
			logger.Debug.Subprocess("$BP_PHP_LIB_DIR = %s", libDir)
			logger.Debug.Break()
		}

		logger.Subprocess("Generating default PHP configuration")
		defaultConfig, buildpackConfig, err := files.WriteConfig(phpLayer.Path, context.CNBPath, PhpIniConfig{
			IncludePath: strings.Join([]string{
				filepath.Join(phpLayer.Path, "lib", "php"),
				filepath.Join(context.WorkingDir, libDir),
			}, string(os.PathListSeparator)),
			ExtensionDir: extensionsDir,
			// TODO: figure out where extensions and zendextensions arrays come from
			// Do we even need to load extensions in the default INI file? Maybe better to simply require that folks add additional INI files?
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Action("Generated %s and %s", defaultConfig, buildpackConfig)
		logger.Break()

		err = environment.Configure(phpLayer, extensionsDir, defaultConfig, []string{
			filepath.Dir(defaultConfig),
			filepath.Dir(buildpackConfig),
			filepath.Join(context.WorkingDir, "php.ini.d"),
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.EnvironmentVariables(phpLayer)

		if phpLayer.Build {
			logger.Debug.Process("PHP layer will be available to other buildpacks during build")
		}
		if phpLayer.Launch {
			logger.Debug.Process("PHP layer will be available at runtime")
		}
		if phpLayer.Cache {
			logger.Debug.Process("PHP layer will be cached")
		}

		return packit.BuildResult{
			Layers: []packit.Layer{phpLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
