package main

import (
	"os"

	packit "github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	phpdist "github.com/paketo-buildpacks/php-dist"
)

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

func main() {
	buildpackYMLParser := phpdist.NewBuildpackYMLParser()
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	entryResolver := draft.NewPlanner()
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := phpdist.NewEnvironment()
	sbomGenerator := Generator{}

	packit.Run(
		phpdist.Detect(
			buildpackYMLParser,
		),
		phpdist.Build(
			entryResolver,
			dependencyManager,
			phpdist.NewPHPFileManager(),
			environment,
			sbomGenerator,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
