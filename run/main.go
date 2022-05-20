package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
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
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := phpdist.NewEnvironment()
	sbomGenerator := Generator{}

	packit.Run(
		phpdist.Detect(),
		phpdist.Build(
			dependencyManager,
			phpdist.NewPHPFileManager(),
			environment,
			sbomGenerator,
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
